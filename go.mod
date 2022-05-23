module github.com/nordicsense/landsat

go 1.18

replace github.com/tensorflow/tensorflow => ../../../Library/go/src/github.com/tensorflow/tensorflow

require (
	github.com/nordicsense/gdal v0.0.0-20220115002029-251cd7760df6
	github.com/tensorflow/tensorflow v2.8.1+incompatible
	github.com/teris-io/cli v1.0.1
	github.com/vardius/progress-go v0.0.0-20210725070013-c85a970b9413
)

require google.golang.org/protobuf v1.28.0 // indirect
