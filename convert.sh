#!/usr/bin/env sh

export VERSION=18classes-7vars

export RESULTS_DIR=/Volumes/Caffeine/Data/Landsat/results/${VERSION}
export ROOT_DIR=/Volumes/Caffeine/Data/Landsat

mkdir -p ${ROOT_DIR}/corrected/training
mkdir -p ${ROOT_DIR}/corrected/prod
landsat correct -v -s -d ${ROOT_DIR}/sources/training -o ${ROOT_DIR}/corrected/training # compress=deflate zlevel=6 predictor=3
landsat correct -v -s -d ${ROOT_DIR}/sources/prod -o ${ROOT_DIR}/corrected/prod # compress=deflate zlevel=6 predictor=3

mkdir -p ${RESULTS_DIR}/trainingdata
landsat field ${ROOT_DIR}/sources/training-coordinates -d ${ROOT_DIR}/corrected/training -o ${RESULTS_DIR}/trainingdata/trainingdata

# Train the model here

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LC08*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=8 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}
done

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LE07*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=7 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}
done

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LT05*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=5 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}
done

