GOPATH=/Users/jafar/Workspaces/go/go-sikopat/.gopath:/Users/jafar/.go

default: build

bind: build
	gomobile bind -javapkg sikopat.api github.com/reekoheek/go-sikopat/api

copy: bind
	cp ./api.aar /Users/jafar/Workspaces/mobile/Sikopat/api/

build:
	gopas build
