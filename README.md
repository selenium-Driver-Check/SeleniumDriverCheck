## SeleniumDriverCheck 
this package is for auto installing chrome driver for Go-selenium

### Usage

```
printLog := true
wd, err := selenium.NewChromeDriverService(SeleniumDriverCheck.AutoDownload_ChromeDriver(printLog), 9515)
```

```
printLog := false
wd, err := selenium.NewChromeDriverService(SeleniumDriverCheck.AutoDownload_ChromeDriver(printLog), 9515)
```
