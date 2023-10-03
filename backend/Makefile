mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(dir $(mkfile_path))
CMDprepareTest != go run tools/prepare_tests/prepare_tests.go -file $(current_dir)test.env
CMDRunTest != go run tools/run_tests/run_tests.go -file $(current_dir)test.env

main:
	echo "select make action"

prepare-test:
	cd ..
	go install github.com/allaboutapps/integresql/cmd/server@latest
	powershell -command '$(CMDprepareTest)'

run-test:
	cd ..
	powershell -command '$(CMDRunTest)'

build-linux-backend:
	powershell -command '$$Env:GOOS = "linux"; $$Env:GOARCH = "amd64"; go build -o $(current_dir)/builds/main .\main.go'

zod-clear:
	powershell -command rm -r .\api.json

zod:
	go run main.go openapi
	npx openapi-zod-client -a .\api.json -o ..\frontend\admin\src\api\zodios\api.ts
	$(MAKE) zod-clear

zod-api:
	go run main.go openapi
	npx openapi-zod-client -a .\api.json -o ..\frontend\admin\src\api\zodios\api.ts
	$(MAKE) zod-clear



build: build-linux-backend

release-pr:
	go run tools/git_release/git_release.go -repo pindamonhangaba/monoboi

dev:
	$(MAKE) zod-api
	go run .\main.go serve --config=./config.yml

