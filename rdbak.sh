#!/bin/sh
export CONFIG=/opt/rdbak/config.json
/opt/rdbak/bin/rdbak backup >>/var/log/rdbak.log 2>&1
#/opt/rdbak/bin/rdbak encrypt-pwd
