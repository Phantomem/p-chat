cd /home/ec2-user/
&& sudo yum install -y git
&& cd /tmp/
&& wget https://go.dev/dl/go1.23.5.linux-arm64.tar.gz
&& export PATH=$PATH:/usr/local/go/bin
&& go version
&& cd /home/ec2-user/
&& git clone https://github.com/Phantomem/p-chat.git
&& cd p-chat
&& go get
&& go build main.go
&& ./main
