package main

import (
	"testing"
	"io/ioutil"
	"strings"
	"fmt"
	"github.com/segmentio/objconv/json"
)

func Test_analysis(t *testing.T) {
	debugMode = true

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	//
	accountID := -1
	startLo := 0

	majsoulRoundData := &majsoulRoundData{accountID: accountID}
	majsoulRoundData.roundData = newRoundData(majsoulRoundData, 0, 0)

	s := struct {
		Message string `json:"message"`
	}{}

	for lo, line := range strings.Split(string(data), "\n")[startLo:] {
		fmt.Println(lo+1)
		json.Unmarshal([]byte(line), &s)

		msg := s.Message
		d := majsoulMessage{}
		if err := json.Unmarshal([]byte(msg), &d); err != nil {
			fmt.Println(err)
			continue
		}

		majsoulRoundData.msg = &d
		majsoulRoundData.originJSON = msg
		if err := majsoulRoundData.analysis(); err != nil {
			fmt.Println("错误：", err)
		}
	}
}
