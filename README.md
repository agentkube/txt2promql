![Agent Kube Thumbnail](https://github.com/user-attachments/assets/d50bad8f-fd3e-4869-9520-8c94b9954a48)

# Txt2PromQL

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/agentkube/txt2promql)](https://goreportcard.com/report/github.com/agentkube/txt2promql)
[![Build Status](https://github.com/agentkube/txt2promql/actions/workflows/publish.yaml/badge.svg)](https://github.com/agentkube/txt2promql/actions)

Convert natural language queries to PromQL with AI-powered understanding. Designed for monitoring democratization and observability workflows.

<!-- **Demo** (insert animated GIF here showing CLI and web interface) -->

## Features

- **Natural Language Interface**: Convert plain English to production-ready PromQL
- **Knowledge Graph Integration**: Understands metric relationships and monitoring best practices
- **Hybrid AI Approach**: Combines LLM capabilities with rule-based validation
- **Multi-Interface Support**:
  - REST API
  - Command Line Interface (CLI)
- **Prometheus Native**:
  - Auto schema discovery
  - Query validation
  - Built-in connection management

## Installation

### From Source

```bash
go install github.com/agentkube/txt2promql@latest
```

###  PromQL queries scenarios


| **Category**                  | **Question**                                                                                                                                           |
|-------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Node Resources**            | Write PromQL statements to query the CPU utilization of each Kubernetes node.                                                                          |
|                               | Write PromQL statements to query the memory usage of the following Kubernetes node: `172.16.xx.xx:9100`.                                               |
|                               | Write PromQL statements to trigger alerts if a node becomes abnormal in the current cluster.                                                           |
|                               | Write PromQL statements to query abnormal nodes.                                                                                                      |
| **Pod Resources**             | Write PromQL statements to trigger alerts if a pod is restarted more than twice within 5 minutes.                                                     |
|                               | Write PromQL statements to query the availability of the following pod: `namespace="arms", deployment="arms-pop-malaysia"`.                             |
|                               | Write PromQL statements to query the pod that has the most exceptions in a Kubernetes cluster.                                                         |
|                               | Write PromQL statements to query the failed jobs.                                                                                                     |
| **Container Resources**       | Write PromQL statements to query the container with the highest CPU utilization in the default namespace.                                              |
|                               | Write PromQL statements to query the memory usage of each container in the following namespace and pod: `namespace="default", pod_name="arms-xtrace"`.  |
|                               | Write PromQL statements to query the top five containers with the highest memory usage.                                                                |
| **Lingjun Dashboard Metrics** | Write PromQL statements to query the GPU utilization of each node on the Lingjun dashboard.                                                            |
|                               | Write PromQL statements to query the GPU utilization of each cluster on the Lingjun dashboard.                                                         |
| **Average Response Time**     | Write PromQL statements to query the average response time of each API operation.                                                                      |
|                               | Write PromQL statements to query the average response time of each API operation in Application A.                                                     |
|                               | Write PromQL statements to query the top 10 API operations with the longest average response time in Application A.                                     |
|                               | Write PromQL statements to query the top 10 applications with the longest average response time.                                                       |
| **Error Rates**               | Write PromQL statements to query the error rate of each API operation in Application A in the previous minute.                                         |
|                               | Write PromQL statements to query the top 10 applications with the highest error rate in ARMS.                                                          |
|                               | Write PromQL statements to query the top 10 API operations with the highest error rate in Application A.                                               |
|                               | Write PromQL statements to query the top five API operations with the highest error rate on the machine (`IP: 195.128.xx.xx`) of Application A in the previous 2 hours. |
| **Number of Calls**           | Write PromQL statements to query the queries per second (QPS) of a Redis database.                                                                    |
|                               | Write PromQL statements to query the QPS of the Dubbo service.                                                                                        |
|                               | Write PromQL statements to query the QPS of each API operation in Application A.                                                                       |
|                               | Write PromQL statements to query the number of API calls for each application in the previous hour and group the calls by machine.                     |
|                               | Write PromQL statements to query the number of calls to each API operation of Application A in the previous 5 minutes in ARMS.                         |
|                               | Write PromQL statements to query the number of calls to the API operations with the `payment/coupon` tag in Application A in the previous 5 minutes.    |
|                               | Write PromQL statements to query the top 10 API operations with the largest number of calls.                                                           |
|                               | Write PromQL statements to query the top five API operations with the largest number of calls in Application A.                                        |
| **Number of Errors**          | Write PromQL statements to query the number of errors for each API operation in the previous 5 minutes.                                               |
|                               | Write PromQL statements to query the total number of HTTP request errors on the machine whose IP address is `10.26.xx.xx` in the previous 5 minutes.   |
|                               | Write PromQL statements to query the API operation with the largest number of errors in the previous hour.                                             |
|                               | Write PromQL statements to query the total number of call errors that occurred on the machine (`IP: 10.26.xx.xx`) and on which the ClothService service is deployed in the previous 10 minutes. |
| **Slow SQL Queries**          | Write PromQL statements to query the slow SQL queries generated in the previous 10 minutes.                                                           |
|                               | Write PromQL statements to query the API operation that causes the largest number of slow SQL queries in Application A in the previous 10 minutes.     |
|                               | Write PromQL statements to query the top 10 API operations that cause the largest number of slow SQL queries in Application A.                         |
|                               | Write PromQL statements to query the slow SQL queries generated in the previous hour.                                                                 |
| **Database Metrics**          | Write PromQL statements to query the API operations that failed to be called on a Redis database in the previous minute.                               |
|                               | Write PromQL statements to query the top five API operations that failed to be called on a MySQL database in the previous minute.                      |
| **HTTP Status Codes**         | Write PromQL statements to count the number of 4xx or 5xx errors.                                                                                     |
|                               | Write PromQL statements to count the number of 400 and 500 errors.                                                                                    |
|                               | Write PromQL statements to query the number of requests for which 4xx is returned for Application A.                                                   |
| **Full Garbage Collections**  | Write PromQL statements to query the number of full GCs occurred in the previous day.                                                                 |
|                               | Write PromQL statements to query the number of full GCs occurred on each machine in the previous hour.                                                 |
|                               | Write PromQL statements to query the machines on which full GCs occurred in Application A.                                                            |
|                               | Write PromQL statements to query the machines on which full GCs occurred more than five times.                                                         |
| **GC Time Consumption**       | Write PromQL statements to query the amount of time consumed by full GCs on each machine.                                                             |
|                               | Write PromQL statements to query the top five machines on which full GCs consume the largest amount of time.                                           |
|                               | Write PromQL statements to query the services in which full GCs consume more than 1 second.                                                           |
| **QPS Increase**              | Write PromQL statements to query the applications whose number of access requests increases within 10 minutes.                                         |
|                               | Write PromQL statements to query the application whose number of access requests most increases in the previous day.                                   |
|                               | Write PromQL statements to query the API operation whose number of access requests most rapidly increases in Application A in the previous week.       |
| **Incremental Metrics**       | Write PromQL statements to monitor the `arms_mysql_requests_error_count` metric and send an alert if the metric value suddenly increases.              |
|                               | Write PromQL statements to monitor the increment of the `jvm_threads_live_threads` metric.                                                             |
| **ARMS Console Metrics**      | Write PromQL statements to check whether the number of errors increases or decreases compared with that of yesterday.                                  |
|                               | Write PromQL statements to query the number of requests that increases or decreases compared with that of the previous hour.                           |
|                               | Write PromQL statements to check whether the number of exceptions increases or decreases compared with that of yesterday.                              |
|                               | Write PromQL statements to check whether the average amount of time consumed by applications increases or decreases compared with that of the previous hour. |
|                               | Write PromQL statements to query the increased or decreased average amount of time consumed by applications.                                           |
|                               | Write PromQL statements to query the applications that are affected by full GCs.                                                                      |
|                               | Write PromQL statements to query the API operations that are affected by full GCs.                                                                    |
|                               | Write PromQL statements to query the applications that involve slow SQL queries.                                                                      |
|                               | Write PromQL statements to query the API operations that cause slow SQL queries.                                                                      |
|                               | Write PromQL statements to query the applications whose number of errors increases.                                                                   |
|                               | Write PromQL statements to query the machines whose number of errors increases.                                                                       |
|                               | Write PromQL statements to query the API operations whose number of errors increases in Application A.                                                |
| **CPU Utilization**           | Write PromQL statements to query the CPU utilization of each machine.                                                                                 |
|                               | Write PromQL statements to query the top five machines with the highest CPU utilization in the previous 5 minutes.                                     |
|                               | Write PromQL statements to query the machines whose CPU utilization exceeds 70% in Application A in the previous 5 minutes.                           |
|                               | Write PromQL statements to query the top five machines whose CPU utilization most rapidly increases in the previous 5 minutes and list the CPU utilization. |
