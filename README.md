# kaobei_go
拷贝漫画的go语言爬虫，加上本地消息队列的实现

kafka的相关命令 

docker exec -it kafka bash
cd /opt/kafka/bin

检查现有 topic
./kafka-topics.sh --bootstrap-server localhost:9092 --list

查看详情（最重要：看 PartitionCount 是否 >=8）：
./kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic comic-chapters

3b4416e9423f:/opt/kafka/bin$ ./kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic comic-chapters
Topic: comic-chapters   TopicId: Pz9D-FPTTtSzXyGLSgYGvA PartitionCount: 8       ReplicationFactor: 1    Configs: 
        Topic: comic-chapters   Partition: 0    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 1    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 2    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 3    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 4    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 5    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 6    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:
        Topic: comic-chapters   Partition: 7    Leader: 1       Replicas: 1     Isr: 1  Elr:    LastKnownElr:

修改partitions数量
./kafka-topics.sh --bootstrap-server localhost:9092 --alter --topic comic-chapters --partitions 8


删除topic
./kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic comic-chapters

确认数量
./kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic comic-chapters
