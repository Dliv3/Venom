echo "build macos x64 admin & agent..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o release/admin_macos_x64 admin/admin.go 
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o release/agent_macos_x64 agent/agent.go 

echo "build linux x64 admin & agent..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o release/admin_linux_x64 admin/admin.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o release/agent_linux_x64 agent/agent.go 

echo "build linux x86 admin & agent..."
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o release/admin_linux_x86 admin/admin.go 
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o release/agent_linux_x86 agent/agent.go

echo "build windows x86 admin & agent..."
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o release/admin.exe admin/admin.go 
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o release/agent.exe agent/agent.go 

# examples for iot:
# arm eabi 5
echo "build arm eabi5 agent..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "-s -w" -o release/agent_arm_eabi5 agent/agent.go
# mips 32 little endian
echo "build mipsel agent..."
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build -ldflags "-s -w" -o release/agent_mipsel_version1 agent/agent.go
