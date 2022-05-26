#!/usr/bin/env sh

mkdir -p /Volumes/Caffeine/Data/Landsat/corrected/training
landsat correct -v -s \
  -d /Volumes/Caffeine/Data/Landsat/sources/training \
  -o /Volumes/Caffeine/Data/Landsat/corrected/training # compress=deflate zlevel=6 predictor=3

mkdir -p /Volumes/Caffeine/Data/Landsat/corrected/prod
landsat correct -v -s \
  -d /Volumes/Caffeine/Data/Landsat/sources/prod \
  -o /Volumes/Caffeine/Data/Landsat/corrected/prod # compress=deflate zlevel=6 predictor=3

mkdir -p /Volumes/Caffeine/Data/Landsat/results/current/trainingdata
landsat field /Volumes/Caffeine/Data/Landsat/sources/training-coordinates \
  -d /Volumes/Caffeine/Data/Landsat/corrected/training \
  -o /Volumes/Caffeine/Data/Landsat/results/current/trainingdata/trainingdata


mkdir -p /Volumes/Caffeine/Data/Landsat/results/current

for TIFFNAME in /Volumes/Caffeine/Data/Landsat/corrected/prod/LC08*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=8 "$TIFFNAME" \
    -m /Volumes/Caffeine/Data/Landsat/results/current/tf.model \
    -o /Volumes/Caffeine/Data/Landsat/results/current
done

for TIFFNAME in /Volumes/Caffeine/Data/Landsat/corrected/prod/LE07*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=7 "$TIFFNAME" \
    -m /Volumes/Caffeine/Data/Landsat/results/current/tf.model \
    -o /Volumes/Caffeine/Data/Landsat/results/current
done

for TIFFNAME in /Volumes/Caffeine/Data/Landsat/corrected/prod/LT05*.tiff; do
  echo $TIFFNAME
  landsat predict -v --id=5 "$TIFFNAME" \
    -m /Volumes/Caffeine/Data/Landsat/results/current/tf.model \
    -o /Volumes/Caffeine/Data/Landsat/results/current
done

