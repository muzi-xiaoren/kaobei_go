# kaobei_go - 拷贝漫画 Go 语言爬虫 + 本地消息队列实现

一个高效、稳定的 **拷贝漫画（kaobei manhua）** 内容爬取工具，使用 Go 语言开发，支持漫画列表抓取、章节解析、图片下载，并集成 **本地消息队列**（基于 Kafka）实现生产者-消费者模式的任务解耦与并发处理。

**核心功能**
- 爬取拷贝漫画网站的作品列表、章节信息、图片资源
- 支持断点续传、多线程/协程并发下载
- 使用消息队列（Kafka）解耦爬取任务与下载任务，避免阻塞，提高吞吐量和容错性
- 一键 Docker Compose 部署（包含 Zookeeper + Kafka），本地快速启动完整环境
- 配置灵活（代理、并发数、下载路径等可通过配置文件或环境变量调整）

**技术栈**
- Go (1.18+) + 标准库 + 第三方包（如 colly/goquery/colly 或 net/http + goquery 解析）
- Kafka (confluentinc/cp-kafka 或 sarama 客户端) 作为本地消息队列
- Docker Compose 一键启动 Kafka 集群

**快速启动**

```bash
# 克隆仓库
git clone https://github.com/muzi-xiaoren/kaobei_go.git
cd kaobei_go

# 启动 Kafka (Zookeeper + Kafka broker)
docker-compose up -d

# 编译 & 运行爬虫
go mod tidy
go run main.go
# 或 go build -o kaobei && ./kaobei

#### kafka的相关命令 

```
docker exec -it kafka bash
cd /opt/kafka/bin

检查现有 topic
./kafka-topics.sh --bootstrap-server localhost:9092 --list

查看详情：
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
```

