#!/bin/sh
#只作用于当前进程,不作用于其创建的子进程
set -e
#$0--Shell本身的文件名 $1--第一个参数 $@--所有参数列表
# allow the container to be started with `--user`
if [ "$1" = 'redis-server' -a "$(id -u)" = '0' ]; then
    sed -i 's/REDIS_PORT/'$REDIS_PORT'/g' /usr/local/etc/redis.conf
    chown -R redis .  #改变当前文件所有者
    exec gosu redis "$0" "$@"  #gosu是sudo轻量级”替代品”
fi

exec "$@"
