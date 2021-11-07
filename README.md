# MysqlEvent

#### mysql binlog 订阅数据处理

    安装需知:
        1. 开启mysql row 模式
        2. 确定 mysqldump 指令在系统 $PATH 内
     
    支持:
        1. 订阅指定字段写入AMQP
        2. 订阅指定字段触发远程回调
        3. 配置热加载
    
    待支持:
        1. 表名正则 支持
        2. 便捷指令 shell 支持
        3. 写入Redis 支持
        4. 进程管理，脱离 supervistor 支持
        
    使用:
        step 1: cd ${GOPATH}/src/mysqlevent
        step 2: make
        step 3: ./bin/mysql-event
    热重载:
        执行: cd ${GOPATH}/src/mysqlevent && ./restart.sh
        
        
    推荐阅读:
    https://zhuanlan.zhihu.com/p/115931103
    https://www.jianshu.com/p/5e6b33d8945f