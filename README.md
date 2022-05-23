# Landsat image analysis suite

## General capabilities

* Landsat L5, L7, L8 atmospheric correction based on stored metadata
* Landsat band aggregation into multi-layer TIFFs
* Collection of training data from mult-layer Landsat TIFF images using the mapping of coordinates to images and classes
* Training and validating a Tensorflow based classifier for Landsat landcover
* Classification of full or partial multi-layer Landsat TIFF images into classification maps

## Dependencies and Requirements

* `libgdal` with `libtiff` support for operating with Landsat images and for creating classification maps
* `libtensorflow` for performing the classification
* `tensorflow` python package for training the classifier


## Building on OSX M1 (arm64 arch)


Add `LIBRARY_PATH` and `DYLD_LIBRARY_PATH` to IntelliJ settings.

```shell
brew install tensorflow
brew install protobuf
brew install swig

export LIBRARY_PATH=/opt/homebrew/lib
export DYLD_LIBRARY_PATH=/opt/homebrew/lib

git clone --branch v2.9.0 https://github.com/tensorflow/tensorflow.git ${GOPATH}/src/github.com/tensorflow/tensorflow
# v2.8.1 has a broken protobuf dependency and needs to be fixed with a cherry pick:
# git cherry-pick --strategy-option=no-renames --no-commit 65a5434

cd ${GOPATH}/src/github.com/tensorflow/tensorflow
go mod init github.com/tensorflow/tensorflow
(cd tensorflow/go/op && go generate)
go mod tidy

go test ./...
```

On OSX arm64 this conflicts with tensorflow from brew, so make sure that `LIBRARY_PATH` and `DYLD_LIBRARY_PATH` are
unset when operating with the python library:

```shell
brew install miniforge

conda create --name landsat python=3.10.2
conda activate landsat
conda install -c apple tensorflow-deps==2.9.0

conda install -c apple pandas tensorflow-deps tensorflow
```

[source](https://towardsdatascience.com/how-to-train-a-classification-model-with-tensorflow-in-10-minutes-fd2b7cfba86)

Saving model in Python and loading it in Go
https://tonytruong.net/running-a-keras-tensorflow-model-in-golang/
