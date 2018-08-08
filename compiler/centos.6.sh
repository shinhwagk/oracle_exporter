tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz

# edit /etc/profile
mkdir -p /root/go/{src,bin}
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export GOBIN=/root/go/bin' >> /etc/profile

yum install -y git
yum install -y oracle-instantclient12.2-basic-12.2.0.1.0-1.x86_64.rpm

# oracle env config
sh -c "echo /usr/lib/oracle/12.2/client64/lib > /etc/ld.so.conf.d/oracle-instantclient.conf"
ldconfig

# oracle env
export LD_LIBRARY_PATH=/usr/lib/oracle/12.2/client64/lib:$LD_LIBRARY_PATH

export PATH=/root/go/bin:$PATH