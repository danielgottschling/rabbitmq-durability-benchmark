while true; do
    timestamp=$(date +%Y-%m-%dT%H:%M:%S)
    cpu_mem=$(top -b -n 1 | grep beam.smp | awk '{print $9","$10}')
    if [[ -n "$cpu_mem" ]]; then
        echo "$timestamp,$cpu_mem" >> rabbitmq_usage.csv
    fi
    sleep 1
done
