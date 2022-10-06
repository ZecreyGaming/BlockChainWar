kill -9 $(lsof -t -i:3250)
git pull
nohup ./start_server.sh > out &
lsof -t -i:3250