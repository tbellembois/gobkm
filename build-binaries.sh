#!/usr/bin/env bash
#
# Developper script to build binaries
#
set -o nounset # Treat unset variables as an error

OUTPUT_DIR="./build"

BINARY_ARMV7_NAME="gobkm-ARMv7"
BINARY_X86_NAME="gobkm-AMD64"

PACKAGE_ARCHIVE_NAME="gobkm.zip"
STATIC_RESOURCES_ARCHIVE_NAME="static.zip"

#BUILD_GOPHERJS_CMD="gopherjs build static/js/gjs-main.go -o static/js/gjs-main.js -m"
BUILD_ARMV7_CMD="env GOOS=linux GOARCH=arm CC=arm-linux-gnueabi-gcc GOARM=7 CGO_ENABLED=1 ENABLE_CGO=1 go build -o $OUTPUT_DIR/$BINARY_ARMV7_NAME ."
BUILD_X86_CMD="go build -o $OUTPUT_DIR/$BINARY_X86_NAME ."

RICE_ARMV7_CMD="rice append --exec $OUTPUT_DIR/$BINARY_ARMV7_NAME"
RICE_X86_CMD="rice append --exec $OUTPUT_DIR/$BINARY_X86_NAME"

echo "-cleaning $OUTPUT_DIR"
rm -Rf $OUTPUT_DIR/*

#echo "-generating JS"
#$BUILD_GOPHERJS_CMD

echo "-building $BINARY_X86_NAME"
$BUILD_X86_CMD

echo "-appending rice data"
$RICE_X86_CMD

#echo "-building $BINARY_ARMV7_NAME"
#$BUILD_ARMV7_CMD

#echo "-appending rice data"
#$RICE_ARMV7_CMD

echo "-building binaries zip"
zip -r $OUTPUT_DIR/$PACKAGE_ARCHIVE_NAME $OUTPUT_DIR

echo "-building static resources zip"
zip -r $OUTPUT_DIR/$STATIC_RESOURCES_ARCHIVE_NAME ./static/manifest ./static/img
