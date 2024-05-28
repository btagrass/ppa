read -p "Please enter the os(Linux / Windows), arch(aMd64 / aRm64) and ver(1.0.0): " os arch ver
if [[ -z $os ]]; then
    os=("linux" "windows")
elif [[ $os =~ "l" ]]; then
    os="linux"
else
    os="windows"
fi
if [[ -z $arch ]]; then
    arch=("amd64" "arm64")
elif [[ $arch == "m" || $arch =~ "am" ]]; then
    arch="amd64"
else
    arch="arm64"
fi
if [[ -z $ver ]]; then
    ver="1.0.0"
fi
name=${PWD##*/}

echo
echo --- Prepare ---
echo [i] Configuring proxy...
go env -w GOPROXY=https://goproxy.cn,direct
echo [i] Upgrading packages...
go get -u
go mod tidy
echo [i] Formatting files...
go fmt -x ./...
echo
echo --- Build ---
echo [i] Target: ${os[@]}/${arch[@]}
if [[ -d web ]]; then
    echo [i] Building web...
    yarn --cwd web install
    yarn --cwd web build
fi
echo [i] Building binaries...
for o in ${os[@]}; do
    for a in ${arch[@]}; do
        echo -e "\t$o-$a"
        rm -rf build/$o-$a/$name-$ver
        if [[ -d gui ]]; then
            if [[ $o == "linux" ]]; then
                CGO_ENABLED=1 GOOS=$o GOARCH=$a go build -ldflags "-s -w" -o build/$o-$a/$name-$ver/$name
            else
                CGO_ENABLED=1 GOOS=$o GOARCH=$a go build -ldflags "-s -w" -o build/$o-$a/$name-$ver/$name.exe
            fi
            cp -r conf build/$o-$a/$name-$ver
        else
            if [[ $o == "linux" ]]; then
                CGO_ENABLED=0 GOOS=$o GOARCH=$a go build -ldflags "-s -w" -o build/$o-$a/$name-$ver/$name
            else
                CGO_ENABLED=0 GOOS=$o GOARCH=$a go build -ldflags "-s -w" -o build/$o-$a/$name-$ver/$name.exe
            fi
            cp -r conf build/$o-$a/$name-$ver
        fi
    done
done
if [[ -n $ver && -f Dockerfile ]]; then
    echo [i] Removing docker...
    docker rm -f $name
    echo [i] Building image...
    docker build -t btagrass/$name:$ver .
    echo [i] Cleaning images...
    docker images -f dangling=true -q | xargs docker rmi -f
fi
echo
echo --- End ---
