package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"log"
	"os/exec"
	"regexp"
	"sopre-tray/icon"
	"strings"
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
}


var serviceRegistry ServicesRegistry

func main() {
	checkServices()
	systray.Run(onReady, onExit)
}

func checkServices() {
	var sRegistry []Service
	for _, v := range serviceArr {
		displayName, serviceName, running := checkService(v)
		fmt.Println(displayName, " ", serviceName, " ", running)

		service := Service{
			DisplayName: displayName,
			ServiceName: serviceName,
			Running:     running,
		}

		sRegistry = append(sRegistry, service)
	}
	serviceRegistry.Services = nil
	serviceRegistry.Services = sRegistry
}

func checkService(servicename string) (string, string, bool) {
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "query", servicename)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	displayName := getDisplayName(servicename)
	serviceName, running := frimselReturn(outb.String())

	fmt.Println(displayName, " ", serviceName, " ", running)

	return displayName, serviceName, running
}

func getDisplayName(servicename string) string {
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "getdisplayname", servicename)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return frimselName(outb.String())
}

func frimselName(out string) string {
	cut := strings.Fields(out)
	re2 := regexp.MustCompile(`_`)
	input2 := re2.ReplaceAllString(strings.Join(cut[5:6], ""), " ")
	cut2 := strings.Fields(input2)
	return strings.Join(cut2[3:], " ")
}

func frimselReturn(out string) (string, bool) {
	cut := strings.Fields(out)
	running := checkRunning(strings.Join(cut[8:9], ""))

	fmt.Println(cut)
	fmt.Println(strings.Join(cut[8:9], ""))



	return strings.Join(cut[1:2], ""), running
}

func checkRunning(running string) bool {
	if running != "1" {
		return true
	}
	return false
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Sopre Windows Services")
	systray.SetTooltip("Start/Stop Sopre Services")

	// AP Services
	mAPQKnowledgeBaseServer := systray.AddMenuItem("AP KnowledgeBase Server", "")
	mAPQDBODBC := systray.AddMenuItem("AP QDB-ODBC", "")
	mAPQServer := systray.AddMenuItem("AP QServer", "")
	mAPQTCE := systray.AddMenuItem("AP QTCE", "")
	mAPQTCEE := systray.AddMenuItem("AP QTCEE", "")

	//Separator
	systray.AddSeparator()

	// EP Services
	mEPQDBODBC := systray.AddMenuItem("EP QDB-ODBC", "")
	mEPQServer := systray.AddMenuItem("EP QServer", "")
	mEPQTCE := systray.AddMenuItem("EP QTCE", "")
	mEPQTCEE := systray.AddMenuItem("EP QTCEE", "")

	//Separator
	systray.AddSeparator()

	//Testentry //Todo remove after, is not used
	mTest := systray.AddMenuItem("Test", "")

	go func() {
		for {
			select {
			case <-mTest.ClickedCh:
				fmt.Println("Checking Services")
				sendNotification("Test", "Running Services Check")
				checkServices()

			case <-mAPQKnowledgeBaseServer.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mAPQDBODBC.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mAPQServer.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mAPQTCE.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mAPQTCEE.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mEPQDBODBC.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mEPQServer.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mEPQTCE.ClickedCh:
				toggleService("VCM_AP_60_QSERVER")

			case <-mEPQTCEE.ClickedCh:

			}
		}
	}()

	/*
		"VCM_AP_60_QKNOWLEDGEBASESERVER",
		"VCM_AP_60_QDBODBC_IS",
		"VCM_AP_60_QSERVER",
		"VCM_AP_60_QTCE",
		"VCM_AP_60_QTCE_EDITOR",
		"VCM_EP_60_QDBODBC_IS",
		"VCM_EP_60_QSERVER",
		"VCM_EP_60_QTCE",
		"VCM_EP_60_QTCE_EDITOR",
	 */

}

func getStateForService(serviceName string) (bool, error){
	for _,v := range serviceRegistry.Services{
		if v.ServiceName == serviceName{
			return v.Running, nil
		}
	}
	return false, errors.New(fmt.Sprintf("Cannot find service for %s", serviceName))
}

func getDisplayNameForService(serviceName string) (string, error){
	for _,v := range serviceRegistry.Services{
		if v.ServiceName == serviceName{
			return v.DisplayName, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Cannot find service for %s", serviceName))
}

func toggleService(serviceName string){

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



func startService(serviceName string){
	cmd := exec.Command("C:\\Windows\\System32\\net.exe", "start", serviceName)
	err := cmd.Start()
	if err != nil {
		fmt.Println("About to fail")
		log.Fatal(err)
	}

	displayName,_:=getDisplayNameForService(serviceName)
	sendNotification("Service Started", fmt.Sprintf("Succesfully started %s", displayName))
}

func stopService(serviceName string){
	cmd := exec.Command("C:\\Windows\\System32\\sc.exe", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	displayName,_:=getDisplayNameForService(serviceName)
	sendNotification("Stopping Started", fmt.Sprintf("Stopping %s", displayName))
}

func updateServiceState(serviceName string, running bool) error {
	for _,v := range serviceRegistry.Services{
		if v.ServiceName == serviceName{
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

func onExit() {
	// clean up here
}
