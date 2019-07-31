package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//RestartManager's executable file name
var whoAmI string
//log to file
var logger *log.Logger

func main() {
	//-----------------logger--------------------
	var logFile *os.File
	_, err := os.Stat("./RestartManager.log")
	if err != nil {
		if os.IsNotExist(err) {
			logFile, err = os.Create("./RestartManager.log")
			if err != nil {
				logger.Fatal(err.Error())
			}
		}else{
			logger.Fatal(err.Error())
		}
	}else{
		logFile, err = os.OpenFile("./RestartManager.log", os.O_RDWR | os.O_APPEND, os.ModePerm)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}
	logger = log.New(logFile, "log:", log.Lshortfile | log.Ltime)
	logger.SetFlags(log.LstdFlags)
	//-----------------logger--------------------

	whoAmI = filepath.Base(os.Args[0])
	logger.Println("whoAmI =", whoAmI)
	appPath := os.Args[1]
	//Read from StdinPipe
	var data string
	_, err = fmt.Scanln(&data)
	if err != nil {
		logger.Printf("错了: %s", err.Error())
	} else {
		if data == "r" || data == "r\n" {
			time.Sleep(3 * time.Second)
			cmd := exec.Command(appPath)
			logger.Println(appPath, "233333")
			err = cmd.Start()
			if err != nil {
				logger.Printf("Cmd.Start(): %s", err.Error())
			}else {
				logger.Println("Restart OK.")
			}
		}else if string(data) == "b"{
			appPath = os.Args[2]
			if len(os.Args) != 4 {
				logger.Fatal("NOT ENOUGH ARGUMENTS!", os.Args)
			}
			logger.Println("os.Arg:", os.Args)
			backupPath := os.Args[1]
			appName := os.Args[3]
			time.Sleep(3 * time.Second)
			err = RollBack(backupPath, appName)
			if err != nil {
				logger.Println(err.Error())
			}else {
				logger.Println("Roll back ok")
				cmd := exec.Command(appPath)
				err = cmd.Start()
				if err != nil {
					logger.Printf("Cmd.Start(): %s", err.Error())
				}else {
					logger.Println("Restart OK.")
				}
			}
		}else {
			logger.Println("data is :", string(data))
		}
	}

	err = logFile.Close()
	if err != nil {
		logger.Println("Close logger: " + err.Error())
	}

	os.Exit(0)
}

//Manually recover BACKUPS if necessary
func RollBack(backupPath, prjName string) error {
	//See if backup exists
	fileInfos, err := ioutil.ReadDir(backupPath)
	if err != nil {
		logger.Println(err.Error())
		return err
	}
	if len(fileInfos) == 0 {
		return errors.New("no backup found")
	}
	return rangeFunc(backupPath, prjName)
}

func rangeFunc(path, prjName string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		logger.Println(err.Error(), path)
		return err
	}
	if !fileInfo.IsDir() {
		if fileInfo.Name() == "RestartManager.log" || fileInfo.Name() == whoAmI {
			return nil
		}
		newPath := strings.Replace(path, "BACKUPS"+string(filepath.Separator)+prjName+string(filepath.Separator), "", 1)
		if strings.Contains(newPath, "BACKUPS") || strings.Contains(newPath, "UPDATES") {
			return nil
		}
		_, err := os.Stat(filepath.Dir(newPath))
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(filepath.Dir(newPath), 0664)
				logger.Println("making:", filepath.Dir(newPath))
				if err != nil {
					return err
				}
			}else {
				return err
			}
		}
		err = os.Rename(path, newPath)
		if err != nil {
			logger.Println(err.Error())
			logger.Println(path, newPath)
			return err
		}
	}else{

		fs, err := ioutil.ReadDir(path)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
		for _, v := range fs {
			if path[len(path)-1:]!=string(filepath.Separator) {
				path += string(filepath.Separator)
			}
			err = rangeFunc(path+v.Name(), prjName)
			if err!=nil {
				logger.Println(err.Error())
				return err
			}
		}
	}
	return nil
}