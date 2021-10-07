package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"log"
	"os/exec"
	"regexp"
	"sopre-tray/icon"
	"strings"
	"time"
)

var serviceArr = [9]string{
	"VCM_AP_60_QKNOWLEDGEBASESERVER",
	"VCM_AP_60_QDBODBC_IS",
	"VCM_AP_60_QSERVER",
	"VCM_AP_60_QTCE",
	"VCM_AP_60_QTCE_EDITOR",
	"VCM_EP_60_QDBODBC_IS",
	"VCM_EP_60_QSERVER",
	"VCM_EP_60_QTCE",
	"VCM_EP_60_QTCE_EDITOR",
}

type ServicesRegistry struct {
	Services []Service
}

type Service struct {
	ServiceName string
	DisplayName string
	Running     bool
	Group       string
	MenuItem    *systray.MenuItem
}

var serviceRegistry ServicesRegistry

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {

	systray.SetIcon(icon.Data)
	systray.SetTitle("Sopre Windows Services")
	systray.SetTooltip("Start/Stop Sopre Services")

	mAll := systray.AddMenuItem("All Services", "")
	mAllStart := mAll.AddSubMenuItem("Start", "")
	mAllStop := mAll.AddSubMenuItem("Stop", "")

	mGroup := systray.AddMenuItem("Groups", "")
	mAPAll := mGroup.AddSubMenuItem("Start All AP Services", "")
	mEPAll := mGroup.AddSubMenuItem("Start All EP Services", "")
	systray.AddSeparator()
	checkServices()

	//Separator
	systray.AddSeparator()
	mConfluence := systray.AddMenuItem("Open Confluence", "")
	//Testentry //Todo remove after, is not used
	mTest := systray.AddMenuItem("Test", "")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for _, v := range serviceRegistry.Services {
			<-v.MenuItem.ClickedCh
			fmt.Println("Clicked on ", v.DisplayName)
			toggleService(v.ServiceName)
		}
	}()

	go func() {
		for {
			select {
			case <-mAllStart.ClickedCh:
				fmt.Println("Start all SOPRE Services")
			case <-mAllStop.ClickedCh:
				fmt.Println("Stop all SOPRE Services")
			case <-mAPAll.ClickedCh:
				fmt.Println("Start all AP Services")
			case <-mEPAll.ClickedCh:
				fmt.Println("Start all EP Services")
			case <-mConfluence.ClickedCh:
				_ = open.Run("https://confluence.sbb.ch/x/TYAGZw")
			case <-mTest.ClickedCh:
				fmt.Println("Checking Services")
				sendNotification("Test", "Running Services Check")
				checkServices()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// clean up here
}

func checkServices() {
	var sRegistry []Service
	for _, v := range serviceArr {
		displayName, serviceName, running, groupName := checkService(v)
		log.Println("Checking Service: ", displayName, " [", serviceName, "] Running: ", running, " Group: ", groupName)

		service := Service{
			DisplayName: displayName,
			ServiceName: serviceName,
			Running:     running,
			Group:       groupName,
			MenuItem:    systray.AddMenuItem(displayName, ""),
		}

		sRegistry = append(sRegistry, service)
	}
	serviceRegistry.Services = nil
	serviceRegistry.Services = sRegistry
}

func checkService(servicename string) (string, string, bool, string) {
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "query", servicename)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	displayName, groupName := getDisplayName(servicename)
	serviceName, running := extractServiceNameAndState(outb.String())
	return displayName, serviceName, running, groupName
}

func getDisplayName(servicename string) (string, string) {
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "getdisplayname", servicename)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return extractDisplayNameAndGroup(outb.String())
}

func extractDisplayNameAndGroup(out string) (string, string) {
	cut := strings.Fields(out)
	re2 := regexp.MustCompile(`_`)
	input2 := re2.ReplaceAllString(strings.Join(cut[5:6], ""), " ")
	cut2 := strings.Fields(input2)

	displayName := strings.Join(cut2[3:], " ")

	regroup := regexp.MustCompile(`AP|EP`)
	group := regroup.FindAllString(displayName, 1)
	groupString := strings.Join(group, "")

	return displayName, groupString
}

func extractServiceNameAndState(out string) (string, bool) {

	restate := regexp.MustCompile(".*STATE.*:.[0-9]")
	state := restate.FindAllString(out, -1)
	stateString := strings.Join(state, "")
	stateString = strings.Replace(stateString, " ", "", -1)
	stateString = strings.Replace(stateString, "STATE:", "", 1)
	running := checkRunning(stateString)

	reservice := regexp.MustCompile("SERVICE_NAME.*:.[_A-Z0-9]*")
	service := reservice.FindAllString(out, -1)
	serviceString := strings.Join(service, "")
	serviceString = strings.Replace(serviceString, " ", "", -1)
	serviceString = strings.Replace(serviceString, "SERVICE_NAME:", "", 1)

	return serviceString, running
}

func checkRunning(running string) bool {
	if running != "1" {
		return true
	}
	return false
}

func getStateForService(serviceName string) (bool, error) {
	for _, v := range serviceRegistry.Services {
		if v.ServiceName == serviceName {
			return v.Running, nil
		}
	}
	return false, errors.New(fmt.Sprintf("Cannot find service for %s", serviceName))
}

func getDisplayNameForService(serviceName string) (string, error) {
	for _, v := range serviceRegistry.Services {
		if v.ServiceName == serviceName {
			return v.DisplayName, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Cannot find service for %s", serviceName))
}

func toggleService(serviceName string) {

	state, err := getStateForService(serviceName)
	if err != nil {
		log.Println(err)
	}
	if state {
		stopService(serviceName)
	} else {
		startService(serviceName)
	}
}

func startService(serviceName string) {
	cmd := exec.Command("C:\\Windows\\System32\\net.exe", "start", serviceName)
	err := cmd.Start()
	if err != nil {
		fmt.Println("About to fail")
		log.Fatal(err)
	}

	for true {
		_, _, state, _ := checkService(serviceName)

		fmt.Println(state)

		if state {
			break
		}

		_ = updateServiceState(serviceName, state)
		time.Sleep(2 * time.Second)
	}

	displayName, _ := getDisplayNameForService(serviceName)
	sendNotification("Service Started", fmt.Sprintf("Succesfully started %s", displayName))
}

func stopService(serviceName string) {
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	for true {
		_, _, state, _ := checkService(serviceName)

		fmt.Println(state)

		if !state {
			break
		}

		_ = updateServiceState(serviceName, state)
		time.Sleep(2 * time.Second)
	}

	displayName, _ := getDisplayNameForService(serviceName)
	sendNotification("Stopping Service", fmt.Sprintf("Stopping %s", displayName))
}

func updateServiceState(serviceName string, running bool) error {
	for _, v := range serviceRegistry.Services {
		if v.ServiceName == serviceName {
			v.Running = running
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Cannot find service for %s", serviceName))
}

func sendNotification(title string, text string) {
	err := beeep.Notify(title, text, "icon/icon.png")
	if err != nil {
		panic(err)
	}
}
