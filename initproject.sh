project=$1
if [ "$project" = "" ]
then
    echo project name is empty, usage shell project_name
    exit 127
fi

myos=$(uname)
case "$myos" in
    "Darwin")
        sed -i "" 's#iris/pkg#'$project'/pkg#g' `grep 'iris/pkg' * -rl`
        #find . -type f -name '*.go' -exec sed -i '' s/iris\/pkg/$project\/pkg/ {} +
        sed -i "" 's#mybinary#'$project'#g' cmd/mybinary/main.go
        sed -i "" 's#mybinary#'$project'#g' Makefile
        ;;
    "Linux")
        sed -i 's#iris/pkg#'$project'/pkg#g' `grep 'iris/pkg' * -rl`
        sed -i 's#mybinary#'$project'#g' cmd/mybinary/main.go
        sed -i 's#mybinary#'$project'#g' Makefile
        ;;
    *)
        echo not support
        exit
        ;;
esac

mv cmd/mybinary cmd/$project
rm -rf .git go.mod go.sum
cd ..
[ -d "iris" ] && mv iris $project
cd $project
go mod init
