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
	"strings"

	"github.com/levigross/grequests"
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

//获取chrome主版本号，用于创建文件夹  >
func GetVersionForCreateFile() string {
	status, chromeVersion := GetChromeVersion()
	if status == true {
		mainVersion := getMajorVersion(chromeVersion)
		return mainVersion
	} else {
		panic("Chrome is not installed.")
	}
}

//获取电脑系统版本 > Get Pc Version
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

//获取当前电脑Chrome版本号 > Get Version for chrome
func GetChromeVersion() (bool, string) {
	if strings.HasPrefix(runtime.GOOS, "win") {
		output, _ := exec.Command("reg", "query", "HKEY_CURRENT_USER\\Software\\Google\\Chrome\\BLBeacon").Output()
		chromeVersion := strings.Split(string(output), "\n")
		for _, v := range chromeVersion {
			if strings.Contains(v, "version") {
				versionId := strings.Split(v, "    ")[3]
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

//获取Chrome Driver 链接  > Get Chrome Driver Url
func GetChromeDriverDownLoadUrl(version string) (string, bool) {
	http_base_url := "http://chromedriver.storage.googleapis.com/"
	//https_base_url := 'https://chromedriver.storage.googleapis.com/'
	platform, architecture := GetPcVersion()
	if platform != "" && architecture != "" {
		return http_base_url + version + "/chromedriver_" + platform + architecture + ".zip", true
	} else {
		return "", false
	}
}

//获取Chromedriver主版本号  > Get Chromedriver major version
func getMajorVersion(version string) string {
	return strings.Split(version, ".")[0]
}

//获取对应的chromedriver版本  > Get matched chrome version
func GetMatchedChromeDriverVersion(version string) (string, error) {
	http_url := "http://chromedriver.storage.googleapis.com"
	//https_url := "https://chromedriver.storage.googleapis.com"
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

//下载Chrome Driver 临时文件  >  Download Chrome Driver Temp File
func Download() (string, string, error) {
	var FilePath, DownLoadedFilePath, ChromeDriverVersion string
	var Err, err error
	var localFile = &os.File{}

	version := GetVersionForCreateFile()
	LocalPath, Err := CheckFile()
	if Err != nil {
		return "", version, Err
	}
	Status, ChromeVersion := GetChromeVersion()
	if !Status {
		panic("Chrome is not installed.")
	}
	if PcPaltform == "windows" {
		ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion[:14])
	} else if PcPaltform == "darwin" {
		ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion)
	} else if PcPaltform == "linux" {
		ChromeDriverVersion, Err = GetMatchedChromeDriverVersion(ChromeVersion)
	}
	if Err != nil {
		return "", version, Err
	}

	DownLoadUrl, _ := GetChromeDriverDownLoadUrl(ChromeDriverVersion)
	// 下载 chromedriver
	resp, err := grequests.Get(DownLoadUrl, nil)
	if err != nil {
		return "", version, Err
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
		return "", version, Err
	}
	defer localFile.Close()
	if _, copyErr := io.Copy(localFile, resp); err != nil {
		return "", version, copyErr
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
		panic("There is no ChromeDriver that matches the current Chrome version. It is recommended to downgrade Chrome and try again!")
	}
	defer zipRead.Close()
	// 遍历压缩包中的每一个文件
	for _, f := range zipRead.File {
		// 打开文件
		rc, rangeErr := f.Open()
		if rangeErr != nil {
			return "", version, rangeErr
		}
		defer rc.Close()

		// 创建目标文件
		if PcPaltform == "windows" {
			CreateFile, creErr := os.Create(LocalPath + "\\" + f.Name)
			os.Rename(LocalPath+"\\"+f.Name, LocalPath+"\\"+version+".exe")
			if creErr != nil {
				return "", version, creErr
			}
			defer CreateFile.Close()
			// 复制文件内容
			if _, copyErr := io.Copy(CreateFile, rc); copyErr != nil {
				return "", version, copyErr
			}
		} else if PcPaltform == "darwin" {
			CreateFile, creErr := os.Create(LocalPath + "/" + f.Name)
			if f.Name == "chromedriver" {
				os.Rename(LocalPath+"/"+f.Name, LocalPath+"/"+version)
				os.Chmod(LocalPath+"/"+version, 0755)
			}
			if creErr != nil {
				return "", version, creErr
			}
			defer CreateFile.Close()
			// 复制文件内容
			if _, copyErr := io.Copy(CreateFile, rc); copyErr != nil {
				return "", version, copyErr
			}
		} else if PcPaltform == "linux" {
			CreateFile, creErr := os.Create(LocalPath + "/" + f.Name)
			if f.Name == "chromedriver" {
				os.Rename(LocalPath+"/"+f.Name, LocalPath+"/"+version)
				os.Chmod(LocalPath+"/"+version, 0755)
			}
			if creErr != nil {
				return "", version, creErr
			}
			defer CreateFile.Close()
			// 复制文件内容
			if _, copyErr := io.Copy(CreateFile, rc); copyErr != nil {
				return "", version, copyErr
			}
		}

	}
	//删除 LocalPath + "\\Tempchromedriver.zip"
	if PcPaltform == "windows" {
		DownLoadedFilePath = LocalPath + "\\" + version + ".exe"
	} else if PcPaltform == "darwin" {
		DownLoadedFilePath = LocalPath + "/" + version
	} else if PcPaltform == "linux" {
		DownLoadedFilePath = LocalPath + "/" + version
	}

	return DownLoadedFilePath, version, nil
}

//删除临时文件  > Delete Temp File
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

//查看driver实例是否存在  >
func CheckDriverInstace() string {
	mainVersion := GetVersionForCreateFile()
	LocalPath, err := CheckFile()
	if PcPaltform == "windows" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "\\" + mainVersion + ".exe")
			if findErr == nil {
				return LocalPath + "\\" + mainVersion + ".exe"
			} else {
				return ""
			}
		}
	} else if PcPaltform == "darwin" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "/" + mainVersion)
			if findErr == nil {
				return LocalPath + "/" + mainVersion
			} else {
				return ""
			}
		}
	} else if PcPaltform == "linux" {
		if err == nil && LocalPath != "" {
			//查看 localPath+"\\"+mainVersion+".exe是否存在
			_, findErr := os.Stat(LocalPath + "/" + mainVersion)
			if findErr == nil {
				return LocalPath + "/" + mainVersion
			} else {
				return ""
			}
		}
	}

	return ""
}

//流程 >  Process  >  Main
func AutoDownload_ChromeDriver(printLog bool) string {
	driverPatch := CheckDriverInstace()
	if driverPatch != "" {
		return driverPatch
	}
	path, mainversion, _ := Download()
	DeleteTemFile(mainversion)
	if printLog {
		fmt.Printf("successful checking chrome driver!")
	}
	return path
}
