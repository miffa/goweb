
#  exe service version debug

echo $(basename $0 )
butedata=$(date +%Y%m%d%H%M%S)
version=${1:-"no_version"}
isdebug=${2:-"no"}

if [ "$isdebug" == "debug" ]
then
    go build -gcflags "-N -l" -ldflags "-X goweb/iriscore/version.Version=$version"
else
    go build  -ldflags "-X goweb/iriscore/version.Version=$version"
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
