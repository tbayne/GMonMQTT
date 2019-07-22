// Copyright 2016 Zack Guo <gizak@icloud.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package main

import (
	"fmt"
	ui "github.com/gizak/termui"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
	//log "github.com/cihub/seelog"
	"github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/alecthomas/kingpin.v2"
)

type dashboardWidgets struct {
	hdr1           *ui.Par
	uptime         *ui.Par
	heapSize       *ui.Par
	maxHeapSize    *ui.Par
	subCount       *ui.Par
	connCountGauge *ui.Gauge
}

var version = "GMonMQTT Version: 0.9.1"

// Values displayed on the dashboard
var gUptime = "0"
var gCurrentHeapSize = "0"
var gMaxHeapSize = "0"
var gSubscriptionCount = "0"
var gConnectionCount = 0
var gMaxConnectionCount = 2
var gCurrentAvgRcvd float64

// Command line argument definition
// --help and --version automatically defined

var app = kingpin.New("GMonMQTT", "A console based MQTT Broker Health Monitor").Version(version)
var broker = app.Flag("broker", "The address of the broker to connect with, defaults to tcp://127.0.0.1:1883").Short('b').Default("tcp://127.0.0.1:1883").String()

//define a function for the default message handler
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

	switch msg.Topic() {

	case "$SYS/broker/uptime":
		var uptimeString = string(msg.Payload())
		s := strings.Split(uptimeString, " ")
		seconds, _ := strconv.Atoi(s[0])
		var durationUp time.Duration = time.Duration(seconds) * time.Second
		gUptime = fmt.Sprintf("%v", durationUp)

	case "$SYS/broker/heap/current":
		gCurrentHeapSize = string(msg.Payload())

	case "$SYS/broker/heap/maximum":
		gMaxHeapSize = string(msg.Payload())

	case "$SYS/broker/subscriptions/count":
		gSubscriptionCount = string(msg.Payload())

	case "$SYS/broker/clients/connected":
		gConnectionCount, _ = strconv.Atoi(string(msg.Payload()))

	case "$SYS/broker/clients/maximum":
		gMaxConnectionCount, _ = strconv.Atoi(string(msg.Payload()))

	case "$SYS/broker/load/messages/received/5min":
		gCurrentAvgRcvd, _ = strconv.ParseFloat(string(msg.Payload()), 3)

	}

	//fmt.Printf("%s: %s\n", msg.Topic(),msg.Payload())
}

func createClientOptions(clientID, raw string) *mqtt.ClientOptions {
	// Setup logging

	//logger, _ := log.LoggerFromConfigAsFile("./logging.xml")
	//logger.Infof("creating ClientOptions")

	opts := mqtt.NewClientOptions().AddBroker(raw)
	opts.SetDefaultPublishHandler(f)
	opts.SetClientID(clientID)

	return opts
}

func subscribe(client mqtt.Client, path string) error {
	if token := client.Subscribe(path, 0, nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func setupSubscriptions(client mqtt.Client) error {

	if err := subscribe(client, "$SYS/broker/uptime"); err != nil {
		return err
	}

	if err := subscribe(client, "$SYS/broker/messages/inflight"); err != nil {
		return err
	}

	if err := subscribe(client, "$SYS/broker/clients/connected"); err != nil {

		return err
	}
	if err := subscribe(client, "$SYS/broker/heap/current"); err != nil {

		return err
	}
	if err := subscribe(client, "$SYS/broker/heap/maximum"); err != nil {

		return err
	}
	if err := subscribe(client, "$SYS/broker/clients/connected"); err != nil {

		return err
	}
	if err := subscribe(client, "$SYS/broker/clients/maximum"); err != nil {

		return err
	}
	if err := subscribe(client, "$SYS/broker/subscriptions/count"); err != nil {

		return err
	}

	if err := subscribe(client, "$SYS/broker/subscriptions/count/5m"); err != nil {

		return err
	}

	return nil
}

/*
heapSize := ui.NewPar("Current Heap Size: ")
    heapSize.X = 1
    heapSize.Y = 5
    heapSize.Height = 1
    heapSize.Width = 50
    heapSize.Border = false
    heapSize.TextFgColor = ui.ColorWhite
    heapSize.TextBgColor = ui.ColorBlack
*/

func createParagraphWidget(initialText string, borderLabel string, vpos int, hpos int, height int, width int, border bool) *ui.Par {

	pw := ui.NewPar(initialText)

	if borderLabel != "" {
		pw.BorderLabel = borderLabel
	}

	pw.Y = vpos
	pw.X = hpos
	pw.Height = height
	pw.Width = width
	pw.Border = border

	return pw
}

func setupWidgets() dashboardWidgets {

	var dbw dashboardWidgets
	//func createParagraphWidget(initialText string, borderLabel string, vpos int, hpos int, height int, width int, border bool)
	//logger.Infof("Setting up widgets")
	dbw.hdr1 = createParagraphWidget("  Press [q] to Quit GMonMQTT", " GMonMQTT ", 0, 0, 3, 35, true)
	dbw.hdr1.TextFgColor = ui.ColorWhite
	dbw.hdr1.BorderFg = ui.ColorCyan

	dbw.uptime = createParagraphWidget("Uptime: ", "", 3, 3, 1, 50, false)
	dbw.uptime.TextFgColor = ui.ColorWhite
	dbw.uptime.TextBgColor = ui.ColorBlack

	dbw.heapSize = createParagraphWidget("Current Heap Size: ", "", 4, 3, 1, 50, false)
	dbw.heapSize.TextFgColor = ui.ColorWhite
	dbw.heapSize.TextBgColor = ui.ColorBlack

	dbw.maxHeapSize = createParagraphWidget("Max Heap Size: ", "", 5, 3, 1, 50, false)
	dbw.maxHeapSize.TextFgColor = ui.ColorWhite
	dbw.maxHeapSize.TextBgColor = ui.ColorBlack

	dbw.subCount = createParagraphWidget("Sub Count: ", "", 6, 3, 1, 50, false)
	dbw.subCount.TextFgColor = ui.ColorWhite
	dbw.subCount.TextBgColor = ui.ColorBlack

	dbw.connCountGauge = ui.NewGauge()
	dbw.connCountGauge.Percent = 0
	dbw.connCountGauge.Width = 70
	dbw.connCountGauge.Height = 3
	dbw.connCountGauge.Y = 8
	dbw.connCountGauge.BorderLabel = " Connections "
	dbw.connCountGauge.Label = strconv.Itoa(gConnectionCount) + "/" + strconv.Itoa(gMaxConnectionCount)
	//g3.LabelAlign = ui.AlignRight
	dbw.connCountGauge.PercentColor = ui.ColorWhite
	dbw.connCountGauge.BarColor = ui.ColorGreen
	return dbw
}

func drawWidgets(dbw *dashboardWidgets) {
	ui.Render(dbw.hdr1)

	dbw.uptime.Text = "MQTT Broker Uptime:." + gUptime
	ui.Render(dbw.uptime)

	dbw.heapSize.Text = "Current Heap Size:.." + gCurrentHeapSize
	ui.Render(dbw.heapSize)

	dbw.maxHeapSize.Text = "Max Heap Size:......" + gMaxHeapSize
	ui.Render(dbw.maxHeapSize)

	dbw.subCount.Text = "Subscription Count:." + gSubscriptionCount
	ui.Render(dbw.subCount)

	var p1 = float64(gConnectionCount) / float64(gMaxConnectionCount)

	dbw.connCountGauge.Percent = int((p1 * 100))
	dbw.connCountGauge.Label = strconv.Itoa(gConnectionCount) + "/" + strconv.Itoa(gMaxConnectionCount)

	ui.Render(dbw.connCountGauge)

}

func main() {

	// Setup some nice short arguments
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')

	kingpin.MustParse(app.Parse(os.Args[1:]))

	runtime.GOMAXPROCS(15)
	dbw := setupWidgets()
	// Setup logging
	//logger, err := log.LoggerFromConfigAsFile("./logging.xml")
	/*
	   if err != nil {
	           fmt.Println(err)
	           os.Exit(1)
	   }
	*/
	opts := createClientOptions("GMonMQTT", *broker)
	client := mqtt.NewClient(opts)

	//logger.Infof("client connecting...")

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//logger.Infof("client connected")
	setupSubscriptions(client)

	// Control-C Handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		client.Disconnect(2)
		ui.StopLoop()
		os.Exit(0)
	}()

	err := ui.Init()

	if err != nil {
		panic(err)
	}
	defer ui.Close()

	// Update dashboard widgets every second
	ui.Handle("/timer/1s", func(e ui.Event) {
		drawWidgets(&dbw)
	})

	// Close on 'q'
	ui.Handle("/sys/kbd/q", func(ui.Event) {

		client.Disconnect(2)
		ui.StopLoop()
		os.Exit(0)
	})
	//logger.Infof("Entering ui loop")
	// Start our processing loop
	go ui.Loop()

	for {
		runtime.Gosched()
	}
}
