#!/bin/bash

PID=`ps -ef | grep mysql-event|grep -v "grep"|awk '{print $2}'`

echo "正在重新加载进程PID: "$PID;

kill -USR1 $PID

if [ $? -eq 0 ]
then
    echo "restart success!"
else
    echo "restart fail!"
fi