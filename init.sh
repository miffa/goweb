project=$1
if [ "$project" = "" ]
then
    echo project name is empty, usage shell project_name
    exit 127
fi

myos=$(uname)
case "$myos" in
    "Darwin")
        #sed -i "" 's#goweb#'$project'#g' `grep goweb * -rl`
        find . -type f -name '*.go' -exec sed -i '' s/goweb/$project/ {} +
        sed -i "" 's#hello#'$project'#g' cmd/goweb/main.go
        ;;
    "Linux")
        sed -i 's#goweb#'$project'#g' `grep goweb * -rl`
        sed -i 's#hello#'$project'#g' cmd/goweb/main.go
        ;;
    *)
        echo not support
        exit
        ;;
esac

mv cmd/goweb cmd/$project
cd ..
[ -d "goweb" ] && mv goweb $project
