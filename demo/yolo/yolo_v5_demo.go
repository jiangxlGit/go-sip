package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "go-sip/logger"
)

type Job struct {
	ID int
}

type JobResult struct {
	JobID int
	Err   error
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numWorkers := 3

	jobChan := make(chan Job)
	resultChan := make(chan JobResult)

	var wg sync.WaitGroup

	// 启动固定数量 worker
	for i := 0; i < numWorkers; i++ {
		go worker(ctx, i, jobChan, resultChan, &wg)
	}

	// 独立收集结果
	go func() {
		for res := range resultChan {
			if res.Err != nil {
				fmt.Printf("Job %d failed: %v\n", res.JobID, res.Err)
				// 根据严重程度，可触发 cancel
				// cancel()
			} else {
				fmt.Printf("Job %d completed successfully\n", res.JobID)
			}
		}
	}()

	// 无限循环生成任务
	jobID := 1
	for {
		select {
		case <-ctx.Done():
			fmt.Println("主程序收到退出信号，停止提交任务")
			close(jobChan) // 不再提交新任务，通知 worker 退出
			wg.Wait()      // 等待已有任务完成
			close(resultChan)
			fmt.Println("所有 worker 已退出，程序完全退出")
			return
		default:
			// 模拟定期产生新任务，例如周期性拉帧或流
			job := Job{ID: jobID}
			jobID++

			wg.Add(1)
			jobChan <- job

			// 可选：限流，防止任务过快提交导致 channel 堆积
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- JobResult, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d 收到退出信号，退出\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d: job channel 已关闭，退出\n", id)
				return
			}

			// 执行任务
			err := task(ctx, id, job)
			results <- JobResult{JobID: job.ID, Err: err}
			wg.Done()
		}
	}
}

func task(ctx context.Context, workerID int, job Job) error {
	fmt.Printf("Worker %d: 开始处理任务 %d\n", workerID, job.ID)

	for {
		select {
		case <-ctx.Done():
			Logger.Info("VedioStreamProcess已退出")
			return nil
		default:
			time.Sleep(2 * time.Second)
			fmt.Printf("Worker %d: 任务 %d执行中...\n", workerID, job.ID)
		}
	}
}