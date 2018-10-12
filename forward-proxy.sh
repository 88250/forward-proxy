#!/bin/bash

cd /root/go/src/github.com/b3log/forward-proxy/ && git checkout . && git pull

proc_num=`ps -fe|grep '/root/go/src/github.com/b3log/forward-proxy/forward-proxy'|grep -v grep|wc -l`
if [ $proc_num -gt 0 ]
then
  killall forward-proxy
  echo 'killed forward-proxy'
  sleep 1
fi

nohup /root/go/src/github.com/b3log/forward-proxy/forward-proxy > /var/log/forward-proxy/forward-proxy.log 2>&1 &

echo 'forward-proxy done'