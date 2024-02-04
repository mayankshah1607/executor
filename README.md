A very simple asynchronous queue exposed over a HTTP API.

## Usage

It exposes a HTTP API at `POST /tasks` that accepts a list of task IDs and the milliseconds to run the task for.
For example: `{"task-1": 15, "task-2": 18}` (`task-1` runs for 15ms, `task-2` for 18ms)

It exposes a HTTP API at `GET /tasks` that returns a list of waiting and running tasks.

Follow the below commands for a demo.

```shell
# Build the binary
make build

# Run the binary
bin/executor

# (In a separate terminal window) Add tasks to the executor
curl --location 'localhost:8080/tasks' \
    --header 'Content-Type: application/json' \
    --data '{"task-1": 10000, "task-2": 15000, "task-3": 9000, "task-4": 8000, "task-5": 15000}'

# watch the internal queue state in real-time
watch 'curl -X GET localhost:8080/tasks | jq'
```

### Tweaking parameters

```shell
# 5 workers, max 50 items allowed in queue
bin/executor -n=5 -q=50
```

