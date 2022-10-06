kill -9 $(lsof -t -i:3250)
git pull
nohup ./run.sh > out &
lsof -t -i:3250