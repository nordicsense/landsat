#!/usr/bin/env sh

export LIBRARY_PATH=/opt/homebrew/lib


export VERSION=v10-16c-10v
export RESULTS_DIR=/Volumes/Caffeine/Data/Landsat/results/${VERSION}
export ROOT_DIR=/Volumes/Caffeine/Data/Landsat


mkdir -p ${ROOT_DIR}/corrected/prod
landsat correct -v -s -d ${ROOT_DIR}/sources/prod -o ${ROOT_DIR}/corrected/prod # compress=deflate zlevel=6 predictor=3

mkdir -p ${ROOT_DIR}/corrected/training
landsat correct -v -s -d ${ROOT_DIR}/sources/training -o ${ROOT_DIR}/corrected/training # compress=deflate zlevel=6 predictor=3

mkdir -p ${RESULTS_DIR}/trainingdata
landsat training ${ROOT_DIR}/sources/training-coordinates -d ${ROOT_DIR}/corrected/training -o ${RESULTS_DIR}/trainingdata/trainingdata

# Train the model here
# conda activate landsat

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LC08*.tiff; do
  echo $TIFFNAME
  landsat predict -v -s --id=8 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}/classification
done

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LE07*.tiff; do
  echo $TIFFNAME
  landsat predict -v -s --id=7 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}/classification
done

for TIFFNAME in ${ROOT_DIR}/corrected/prod/LT05*.tiff; do
  echo $TIFFNAME
  landsat predict -v -s --id=5 "$TIFFNAME" -m ${RESULTS_DIR}/tf.model -o ${RESULTS_DIR}/classification
done

# Trim

for TIFFNAME in ${RESULTS_DIR}/classification/*.tiff; do
  echo $TIFFNAME
  landsat trim -v -s "$TIFFNAME" -o ${RESULTS_DIR}
done


# Filter images

for TIFFNAME in ${RESULTS_DIR}/trimmed/*.tiff; do
  echo $TIFFNAME
  landsat filter -v -s 5x5 "$TIFFNAME" -o ${RESULTS_DIR}
done

# Stats

for TIFFNAME in ${RESULTS_DIR}/5x5/*.tiff; do
  landsat stats "$TIFFNAME"
done


# Diff

mkdir -p ${RESULTS_DIR}/diff

cd ${RESULTS_DIR}/5x5 || return

landsat change -o ${RESULTS_DIR}/diff \
  LT05_L2SP_188013_19850709_20200918_02_T1_SR.tiff,LT05_L2SP_188012_19850709_20200918_02_T1_SR.tiff \
  LC08_L2SP_187013_20170710_20200903_02_T1_SR.tiff,LC08_L2SP_187012_20170710_20200903_02_T1_SR.tiff

landsat change -o ${RESULTS_DIR}/diff \
  LT05_L2SP_188013_19850709_20200918_02_T1_SR.tiff,LT05_L2SP_188012_19850709_20200918_02_T1_SR.tiff \
  LC08_L2SP_187013_20210705_20210713_02_T1_SR.tiff,LC08_L2SP_187012_20210705_20210713_02_T1_SR.tiff


cd - || return
