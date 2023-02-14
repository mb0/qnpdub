#!/bin/bash

# Bash error handling: exit immediately if a command exits with a non-zero status
set -e
# Set globbing options so the ./queue/* does not return the pattern if no files were found
shopt -s nullglob dotglob

# switch statement with two cases $1 is the first argument
case $1 in
'run')
	# if the first arg is 'run' ensure some folders exist
	mkdir -p ./queue/done ./queue/fail
	# while true loop
	while [ 1 ]; do
		# local var with count of jobs
		job_count=0
		for job in ./queue/*; do
			# for each job file name in the queue folder that is not a directory
			if [[ -f $job && ! -d $job ]]; then
				# increment job count
				job_count=$((job_count+1))
				echo
				echo "start $(date +%H:%M:%S): $job"
				# run the job and wrapped in the bash time command
				# if the command succeeds move the job file to done folder
				time bash "$job" && mv "$job" ./queue/done && echo "end: $job"
				# if the job file is still there move it to fail folder
				if [ -f $job ]; then
					mv "$job" ./queue/fail && echo "fail: $job"
				fi
			fi
		done
		# if we have not found any job file exit otherwise continue loop.
		# (the glob on queue might me outdated and we need to re-read the folder content)
		if [ $job_count -eq 0 ]; then
			echo "no more jobs found"
			exit 0
		fi
	done;;
*)
	# if the first arg is not 'run' create a new job name using the current time
	job="./queue/$(date +%m%d%H%M%S).job"
	echo "enqueue: $@ as $job"
	# make sure the folder exists and write a script to that file using cat and bash here string
	# the blender output is piped to a log file in the same folder
	mkdir -p ./queue
	cat <<EOF >> $job;;
echo "render $@"
blender -b '$1' ${@:2} -a > '$1.log'
EOF
esac

# The idea behind this script is that you can drop it anywhere and it will setup the queue folder,
# that can be inspected with a normal 'ls' command: 'ls queue/done'. The blender logfiles can be
# inspected with a 'watch' command: 'watch tail -n 3 path/to/blenderfile.log' for the currect frame.

# When writing this in go we could query the total frames from the blender file and print prettier
# progress for extra credit :)
