package SeleniumDriverCheck

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
)

type ListBucketResult struct {
	XMLName  xml.Name   `xml:"ListBucketResult"`
	Contents []Contents `xml:"Contents"`
}
type Contents struct {
	Key string `xml:"Key"`
}

var fileName = "SeleniumDriverPath"
var PcPaltform = runtime.GOOS

func CheckFile() (string, error) {
	//检测用户系统是windows还是mac
	if runtime.GOOS == "windows" {
		//检测当前用户电脑用户名
		u, err := user.Current()
		if err == nil {
			//查看C:\Users\u.Username\AppData\Local下有哪些文件
			dir := fmt.Sprintf(`C:\Users\%s\AppData\Local`, GetRealName(u.Username))
			files, readErr := ioutil.ReadDir(dir)
			if readErr != nil {
				return "", readErr
			}
			for _, checkfile := range files {
				if checkfile.Name() == fileName {
					return filepath.Join(dir, checkfile.Name()), nil
				}
			}
			createErr := os.Mkdir(filepath.Join(dir, fileName), os.ModePerm)
			if createErr == nil {
				return filepath.Join(dir, fileName), nil
			}
		}
	} else if runtime.GOOS == "darwin" {
		//检测当前用户电脑用户名
		u, err := user.Current()
		if err == nil {
			//查看/Users/u.Username/Library/Application Support下有哪些文件
			dir := fmt.Sprintf(`/Users/%s/`, GetRealName(u.Username))
			files, readErr := ioutil.ReadDir(dir)
			if readErr != nil {
				return "", readErr
			}
			for _, checkfile := range files {
				if checkfile.Name() == fileName {
					return filepath.Join(dir, checkfile.Name()), nil
				}
			}
			createErr := os.Mkdir(filepath.Join(dir, fileName), os.ModePerm)
			if createErr == nil {
				return filepath.Join(dir, fileName), nil
			}
		}
	} else if runtime.GOOS == "linux" {
		//检测当前用户电脑用户名
		//查看/home/u.Username/.local/share下有哪些文件
		dir := `/home/`
		files, readErr := ioutil.ReadDir(dir)
		if readErr != nil {
			return "", readErr
		}
		for _, checkfile := range files {
			if checkfile.Name() == fileName {
				return filepath.Join(dir, checkfile.Name()), nil
			}
		}
		createErr := os.Mkdir(filepath.Join(dir, fileName), os.ModePerm)
		if createErr == nil {
			return filepath.Join(dir, fileName), nil
		}
	}

	return "", nil
}
func GetRealName(fullName string) string {
	if strings.Contains(fullName, "\\") {
		realName := strings.Split(fullName, "\\")[1]
		return realName
	}
	if strings.Contains(fullName, "/") {
		realName := strings.Split(fullName, "/")[1]
		return realName
	}
	return fullName
}

// 获取chrome主版本号，用于创建文件夹  >
func GetVersionForCreateFile() string {
	status, chromeVersion := GetChromeVersion()
	if status == true {
		mainVersion := getMajorVersion(chromeVersion)
		return mainVersion
	} else {
		panic("Chrome is not installed.")
	}
}

// 获取电脑系统版本 > Get Pc Version
func GetPcVersion() (string, string) {
	var platform, architecture string
	if strings.HasPrefix(runtime.GOOS, "win") {
		platform = "win"
		architecture = "32"
	} else if strings.HasPrefix(runtime.GOOS, "darwin") {
		platform = "mac"
		architecture = "64"
	} else if strings.HasPrefix(runtime.GOOS, "linux") {
		platform = "linux"
		architecture = "64"
	} else {
		return "", ""
	}
	return platform, architecture
}

// 获取当前电脑Chrome版本号 > Get Version for chrome
func GetChromeVersion() (bool, string) {
	if strings.HasPrefix(runtime.GOOS, "win") {
		output, _ := exec.Command("reg", "query", "HKEY_CURRENT_USER\\Software\\Google\\Chrome\\BLBeacon").Output()
		chromeVersion := strings.Split(string(output), "\n")
		for _, v := range chromeVersion {
			if strings.Contains(v, "version") {
				temp_versionId := strings.Split(v, "    ")[3]
				versionId := strings.Split(temp_versionId, "\r")[0]
				return true, versionId
			}
		}
		return false, ""
	} else if strings.HasPrefix(runtime.GOOS, "darwin") {
		out, err := exec.Command("sh", "-c", `"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --version`).Output()
		if err != nil {
			return false, ""
		}
		version := strings.Split(string(out), " ")[2]
		return true, version
	} else if strings.HasPrefix(runtime.GOOS, "linux") {
		out, err := exec.Command("sh", "-c", `google-chrome --version`).Output()
		if err != nil {
			return false, ""
		}
		version := strings.Split(string(out), " ")[2]
		return true, version
	}
	return false, "Can't Found"
}

// 获取Chrome Driver 链接  > Get Chrome Driver Url
func GetChromeDriverDownLoadUrl(version string) (string, bool) {
	http_base_url := "http://chromedriver.storage.googleapis.com/"
	platform, architecture := GetPcVersion()
	if platform != "" && architecture != "" {
		return http_base_url + version + "/chromedriver_" + platform + architecture + ".zip", true
	} else {
		return "", false
	}
}

func GetNewChromeDriverDownLoadUrl(version string) (string, bool) {
	http_base_url := "https://edgedl.me.gvt1.com/edgedl/chrome/chrome-for-testing/"
	platform, architecture := GetPcVersion()
	if platform != "" && architecture != "" {
		return http_base_url + version + "/" + platform + architecture + "/chromedriver-" + platform + architecture + ".zip", true
	} else {
		return "", false
	}
}

// 获取Chromedriver主版本号  > Get Chromedriver major version
func getMajorVersion(version string) string {
	return strings.Split(version, ".")[0]
}

// 获取对应的chromedriver版本  > Get matched chrome version
func GetMatchedChromeDriverVersion(version string) (string, error) {
	http_url := "http://chromedriver.storage.googleapis.com"
	res, err := grequests.Get(http_url, nil)
	//res是xml,解析
	if err == nil {
		versionList := &ListBucketResult{}
		xml.Unmarshal([]byte(res.String()), &versionList)
		for _, k := range versionList.Contents {
			if strings.Index(k.Key, getMajorVersion(version)+".") == 0 {
				return strings.Split(k.Key, "/")[0], nil
			}
		}
	}
	return "", err
}

func GetNewDriverVersion(mainVersion string) (string, error) {
	http_url := "https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone.json"
	res, err := grequests.Get(http_url, nil)
	if err == nil {
		DriverVersion := gjson.Get(res.String(), "milestones."+mainVersion+".version")
		return DriverVersion.Str, nil
	}
	return "", err
}

// 下载Chrome Driver 临时文件  >  Download Chrome Driver Temp File
func Download(mainVersion string) (string, string, error) {
	var FilePath, DownLoadedFilePath, ChromeDriverVersion, DownLoadUrl string
	var Err, err error
	var localFile = &os.File{}

	mainVersionNumber, err := strconv.Atoi(mainVersion)
	LocalPath, Err := CheckFile()
	if Err != nil {
		return "", mainVersion, Err
	}
	Status, ChromeVersion := GetChromeVersion()
	// Remove below, only for testing purposes
	//ChromeVersion = "114.0.5735.90"
	if !Status {
		panic("Chrome is not installed.")
	}
	if mainVersionNumber < 115 {
		if PcPaltform == "windows" {
			ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion)
		} else if PcPaltform == "darwin" {
			ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion)
		} else if PcPaltform == "linux" {
			ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion)
		}
		if Err != nil {
			return "", mainVersion, Err
		}
		DownLoadUrl, _ = GetChromeDriverDownLoadUrl(ChromeDriverVersion)
	} else {
		ChromeDriverVersion, Err := GetNewDriverVersion(mainVersion)
		if Err != nil {
			return "", mainVersion, Err
		}
		DownLoadUrl, _ = GetNewChromeDriverDownLoadUrl(ChromeDriverVersion)
	}

	// 下载 chromedriver
	resp, err := grequests.Get(DownLoadUrl, nil)
	if err != nil {
		return "", mainVersion, Err
	}
	defer resp.Close()
	// 保存文件到本地
	if PcPaltform == "windows" {
		localFile, err = os.Create(LocalPath + "\\Tempchromedriver.zip")

	} else if PcPaltform == "darwin" {
		localFile, err = os.Create(LocalPath + "/Tempchromedriver.zip")
	} else if PcPaltform == "linux" {
		localFile, err = os.Create(LocalPath + "/Tempchromedriver.zip")
	}
	if err != nil {
		return "", mainVersion, Err
	}
	defer localFile.Close()
	if _, copyErr := io.Copy(localFile, resp); err != nil {
		return "", mainVersion, copyErr
	}
	if PcPaltform == "windows" {
		FilePath = LocalPath + "\\Tempchromedriver.zip"
	} else if PcPaltform == "darwin" {
		FilePath = LocalPath + "/Tempchromedriver.zip"
	} else if PcPaltform == "linux" {
		FilePath = LocalPath + "/Tempchromedriver.zip"
	}
	//将filePath中的chromedriver.exe文件解压到当前文件夹
	// 打开压缩包文件
	zipRead, err := zip.OpenReader(FilePath)
	if err != nil {
		panic(err)
	}
	defer zipRead.Close()
	// 遍历压缩包中的每一个文件
	for _, f := range zipRead.File {
		// 打开文件
		if filepath.Ext(f.Name) == ".chromedriver" {
			continue
		}
		rc, rangeErr := f.Open()
		if rangeErr != nil {
			return "", mainVersion, rangeErr
		}
		defer rc.Close()

		// 创建目标文件
		if PcPaltform == "windows" {

			// Skip chromedriver files
			if filepath.Ext(f.Name) == ".chromedriver" {
				continue
			}

			// Open file in zip archive
			rc, err := f.Open()
			if err != nil {
				return "", mainVersion, err
			}
			defer rc.Close()

			// Extract the folder and file names
			parts := strings.Split(f.Name, "/")
			folderPath := ""
			if len(parts) > 1 {
				folderPath = strings.Join(parts[:len(parts)-1], string(os.PathSeparator))
			}

			// Create directories if needed
			if folderPath != "" {
				if err := os.MkdirAll(LocalPath+string(os.PathSeparator)+folderPath, os.ModePerm); err != nil {
					return "", mainVersion, err
				}
			}

			// Create target file
			filePath := filepath.Join(LocalPath, f.Name)
			if PcPaltform == "windows" {
				filePath = strings.Replace(filePath, "/", "\\", -1)
			}

			createFile, creErr := os.Create(filePath)
			if creErr != nil {
				return "", mainVersion, creErr
			}

			// Copy file content
			if _, copyErr := io.Copy(createFile, rc); copyErr != nil {
				createFile.Close()
				return "", mainVersion, copyErr
			}

			// Close the file
			createFile.Close()

			// Move the copied file one level up
			newFilePath := filepath.Join(LocalPath, filepath.Base(filePath))
			if err := os.Rename(filePath, newFilePath); err != nil {
				return "", mainVersion, err
			}

			dirFilePath := filepath.Join(LocalPath, folderPath)
			if err := os.Remove(dirFilePath); err != nil {
				continue
			}

		} else if PcPaltform == "darwin" {
			CreateFile, creErr := os.Create(LocalPath + "/" + f.Name)
			if f.Name == "chromedriver" {
				os.Rename(LocalPath+"/"+f.Name, LocalPath+"/"+mainVersion)
				os.Chmod(LocalPath+"/"+mainVersion, 0755)
			}
			if creErr != nil {
				return "", mainVersion, creErr
			}
			defer CreateFile.Close()
			// 复制文件内容
			if _, copyErr := io.Copy(CreateFile, rc); copyErr != nil {
				return "", mainVersion, copyErr
			}
		} else if PcPaltform == "linux" {
			CreateFile, creErr := os.Create(LocalPath + "/" + f.Name)
			if f.Name == "chromedriver" {
				os.Rename(LocalPath+"/"+f.Name, LocalPath+"/"+mainVersion)
				os.Chmod(LocalPath+"/"+mainVersion, 0755)
			}
			if creErr != nil {
				return "", mainVersion, creErr
			}
			defer CreateFile.Close()
			// 复制文件内容
			if _, copyErr := io.Copy(CreateFile, rc); copyErr != nil {
				return "", mainVersion, copyErr
			}
		}

	}
	//删除 LocalPath + "\\Tempchromedriver.zip"
	if PcPaltform == "windows" {
		DownLoadedFilePath = LocalPath + "\\" + mainVersion + ".exe"
	} else if PcPaltform == "darwin" {
		DownLoadedFilePath = LocalPath + "/" + mainVersion
	} else if PcPaltform == "linux" {
		DownLoadedFilePath = LocalPath + "/" + mainVersion
	}

	return DownLoadedFilePath, mainVersion, nil
}

// 删除临时文件  > Delete Temp File
func DeleteTemFile(version string) {
	LocalPath, _ := CheckFile()
	if PcPaltform == "windows" {
		os.Remove(LocalPath + "\\Tempchromedriver.zip")
		os.Remove(LocalPath + "\\LICENSE.chromedriver")
		os.Rename(LocalPath+"\\"+"chromedriver"+".exe", LocalPath+"\\"+version+".exe")
	} else if PcPaltform == "darwin" {
		os.Remove(LocalPath + "/Tempchromedriver.zip")
		os.Remove(LocalPath + "/LICENSE.chromedriver")
	} else if PcPaltform == "linux" {
		os.Remove(LocalPath + "/Tempchromedriver.zip")
		os.Remove(LocalPath + "/LICENSE.chromedriver")
	}

}

// 查看driver实例是否存在  >
func CheckDriverInstace() (string, string) {
	mainVersion := GetVersionForCreateFile()
	LocalPath, err := CheckFile()
	if PcPaltform == "windows" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "\\" + mainVersion + ".exe")
			if findErr == nil {
				return LocalPath + "\\" + mainVersion + ".exe", mainVersion
			} else {
				return "", mainVersion
			}
		}
	} else if PcPaltform == "darwin" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "/" + mainVersion)
			if findErr == nil {
				return LocalPath + "/" + mainVersion, mainVersion
			} else {
				return "", mainVersion
			}
		}
	} else if PcPaltform == "linux" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "/" + mainVersion)
			if findErr == nil {
				return LocalPath + "/" + mainVersion, mainVersion
			} else {
				return "", mainVersion
			}
		}
	}

	return "", mainVersion
}

// 流程 >  Process  >  Main
func AutoDownload_ChromeDriver(printLog bool) string {
	driverPatch, mainVersion := CheckDriverInstace()
	if driverPatch != "" {
		return driverPatch
	}
	// Remove below, only for testing purposes
	//mainVersion = "114"
	path, mainversion, _ := Download(mainVersion)
	DeleteTemFile(mainversion)
	if printLog {
		fmt.Printf("successful checking chrome driver!")
	}
	return path
}

func TestNew() (string, error) {
	test_result, err := GetNewDriverVersion("115")
	if err == nil {
		return test_result, err
	}
	return "", err
}
