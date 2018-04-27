#!/bin/bash
unzip -o instantclient-basic-linux.x64-12.2.0.1.0.zip
docker build -t wex/oracle_expoter .