
#  exe service version debug

echo $(basename $0 )
butedata=$(date +%Y%m%d%H%M%S)
service=${1:-"devops"}
version=${2:-"no_version"}
isdebug=${3:-"no"}

if [ "$isdebug" == "debug" ]
then
    go build -gcflags "-N -l" -ldflags "-X iris/pkg/version.Version=$version -X iris/pkg/version.Service=$service"
else
    go build  -ldflags "-X iris/pkg/version.Version=$version -X iris/pkg/version.Service=$service"
fi
if [ $? -eq 0 ]
then
    echo build ok

else
    echo build err
    exit 127
fi
tempdata=$(pwd)
execname=${tempdata##*/}
cp $execname $execname.$version.$butedata
echo $execname.$version.$butedata   is build ok
