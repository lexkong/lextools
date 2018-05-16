package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

func GetLocalAddressOld() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetLocalAddress() string {
	// get available network interfaces for
	// this machine
	interfaces, err := net.Interfaces()

	if err != nil {
		return GetLocalAddressOld()
	}

	ips := make(map[string]string)

	for _, i := range interfaces {
		byNameInterface, err := net.InterfaceByName(i.Name)
		if err != nil {
			return GetLocalAddressOld()
		}

		addresses, err := byNameInterface.Addrs()
		if len(addresses) > 0 {
			ips[i.Name] = strings.Split(addresses[0].String(), "/")[0]
		}
	}
	for _, ifname := range []string{"br1", "bond1", "eth1", "eth0"} {
		if value, ok := ips[ifname]; ok {
			return value
		}
	}

	return GetLocalAddressOld()
}

func FileExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

func Mkdir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

func EnsureDirExists(dir string) (err error) {
	f, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			//如果不存在，创建
			return os.MkdirAll(dir, os.FileMode(0755))
		} else {
			return err
		}
	}

	if !f.IsDir() {
		//已存在，但是不是文件夹
		return fmt.Errorf("path %s is exist,but not dir", dir)
	}

	return nil
}

func DecompressFile(compressedFile string, decompressDir string) error {
	// check the file
	if !strings.HasSuffix(compressedFile, "tar.gz") {
		return fmt.Errorf("%s is not a tar.gz file", compressedFile)
	}

	f, err := os.Stat(compressedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is not exist", compressedFile)
		} else {
			return fmt.Errorf("unknow error when get info of %s", compressedFile)
		}
	}

	if f.IsDir() {
		return fmt.Errorf("%s is a directory", compressedFile)
	}

	// ensure dest dir exist
	if err := EnsureDirExists(decompressDir); err != nil {
		return fmt.Errorf("ensure target dir:%s exist failed:%v", decompressDir, err)
	}

	var args []string
	args = append(args, "zxf")
	args = append(args, compressedFile)
	args = append(args, "-C")
	args = append(args, decompressDir)
	cmd := exec.Command("tar", args...)
	/*
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("run tar command, failed:%v, arguments:%v", err, args)
		}
	*/
	_, err = cmd.CombinedOutput()

	return nil
}

func RemoveFiles(paths []string) error {
	for _, p := range paths {
		if p == "" {
			continue
		}

		cmd := exec.Command("rm", "-rf", p)
		if _, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("exec rm -rf %s failed:%v", p, err)
		}
	}

	return nil
}

func IsYamlString(yamlString string) bool {
	_, err := yaml.YAMLToJSON([]byte(yamlString))
	if err != nil {
		return false
	}

	return true
}

func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 50)
	suffix = strings.ToUpper(suffix)                                                     //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

func GenId() (string, error) {
	iw, err := goSnowFlake.NewIdWorker(1)
	if err != nil {
		return "", err
	}

	id, err := iw.NextId()
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

func CombineRequestErr(resp gorequest.Response, body string, errs []error) error {
	var e, sep string
	if len(errs) > 0 {
		for _, err := range errs {
			e = sep + err.Error()
			sep = "\n"
		}
		return fmt.Errorf("%v", e)
	}

	if resp == nil {
		return fmt.Errorf("response is nil")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s", body)
	}

	return nil
}

func MarshalList(strList []string) (string, error) {
	strJson := ""

	if len(strList) != 0 {
		strByte, err := json.Marshal(strList)
		if err != nil {
			return strJson, err
		}

		strJson = string(strByte)
	}

	return strJson, nil
}

func UnmarshalList(strJsonPtr *string) ([]string, error) {
	var strList []string
	if strJsonPtr != nil && *(strJsonPtr) != "" {
		err := json.Unmarshal([]byte(*strJsonPtr), &strList)
		if err != nil {
			return strList, err
		}
	}

	return strList, nil
}

func DelFromSlice(slice []string, elems ...string) []string {
	isInElems := make(map[string]bool)
	for _, elem := range elems {
		isInElems[elem] = true
	}
	w := 0
	for _, elem := range slice {
		if !isInElems[elem] {
			slice[w] = elem
			w += 1
		}
	}
	return slice[:w]
}

func ConvertTime(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05")
	if s == "0001-01-01 00:00:00" {
		return ""
	}

	return s
}

func Duplicate(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

func IsNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func GenAlnumShortId() string {
	alSet := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())

	id, _ := shortid.Generate()
	id = strings.Replace(id, "_", string(alSet[rand.Intn(len(alSet)-1)]), -1) 
	id = strings.Replace(id, "-", string(alSet[rand.Intn(len(alSet)-1)]), -1) 

	return strings.ToLower(id)
}  

