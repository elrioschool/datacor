#!/bin/bash
# Usage: ./deploy.sh
# Description: Deploy the function to GCP
# Prerequisites: gcloud CLI installed and configured
set -xe

FUNCTION_NAME="donations-by-student-report"

gcloud functions deploy $FUNCTION_NAME --source .
