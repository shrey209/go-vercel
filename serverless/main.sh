#!/bin/bash

# AWS Creds
export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""
export AWS_REGION=""

#git url of user
if [ -z "$GIT_REPOSITORY__URL" ]; then
  echo " Error: GIT_REPOSITORY__URL is not set."
  echo "Usage: GIT_REPOSITORY__URL='your-repo-url' CODE_PATH='your-folder' bash test.sh"
  exit 1
fi

# code path 
if [ -z "$CODE_PATH" ]; then
  echo " Error: CODE_PATH is not provided."
  exit 1
fi

# Clone the repo to code-storage
git clone "$GIT_REPOSITORY__URL" /home/app/code-storage

# Mv to the code-storage
cd /home/app/code-storage || { echo " Error: Failed to enter /home/app/code-storage"; exit 1; }

# Set the target path where React/Vite code exists
TARGET_PATH="$(pwd)/$CODE_PATH"

# Check if the directory exists
if [ ! -d "$TARGET_PATH" ]; then
  echo "Error: $TARGET_PATH does not exist."
  exit 1
fi

# Move into the target directory
cd "$TARGET_PATH" || exit

# npm install
echo " Running npm install..."
npm install
if [ $? -ne 0 ]; then
  echo "Error: npm install failed."
  exit 1
fi

# Build 
echo "⚙️ Running npm run build..."
npm run build
if [ $? -ne 0 ]; then
  echo " Error: npm run build failed."
  exit 1
fi

# dist folder
DIST_FOLDER="$TARGET_PATH/dist"

# move to dist folder
if [ ! -d "$DIST_FOLDER" ]; then
  echo " Error: $DIST_FOLDER does not exist. Build the project before uploading."
  exit 1
fi

# S3 bucket config
S3_BUCKET=""
S3_FOLDER=""

# upload to s3
echo " Uploading dist folder to S3..."
aws s3 cp "$DIST_FOLDER" "s3://$S3_BUCKET/$S3_FOLDER/" --recursive

# Check foru upload status
if [ $? -eq 0 ]; then
    echo " Upload completed successfully to s3://$S3_BUCKET/$S3_FOLDER/"
else
    echo " Upload failed. Check your AWS credentials and bucket permissions."
    exit 1
fi
