# Golang-pipeline

Simple console pipeline that reads integers from stdin, filters them, and periodically prints buffered results.

## 
- Reads numbers from stdin (type `exit` to stop)
- Filters out negative numbers
- Filters out zero and numbers not divisible by 3
- Buffers accepted numbers and prints them every 30 seconds

## Makefile helpers
```bash
make run
make docker-build
make docker-run
```
