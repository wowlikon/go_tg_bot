#!/bin/bash
go mod init github.com/wowlikon/go_tg_bot
go mod tidy
go build && ./go_tg_bot