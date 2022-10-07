My golang web app 1
===
This is a twitter-like web application.
Users can create their own account by their e-mail.
After Sign-up, they can post their articles and read article written by someone they followed.
This app used **golang**, **postgresql**, **neo4j** and **redis** for implementation.
Also used **nginx** to be web server.


## Beginners Guide

If you are a total beginner to this, start here!

If you want to run this app on your loacl or VMs:
1. Make sure postgresql, neo4j, redis has been installed in your environment. If not, you can run env-setting.sh in myapp1/goapp/ to install them.
    ```
    sh env-setiting.sh
    ```
2. If you want your app to run behind a web-server, make sure nginx or any web-server you want has been installed in your environment. This project used nginx to be the web server. You can find default.conf in myapp1/nginx/. Copy and paste it to /etc/nginx/conf.d/ after you install nginx. And modify default.conf in /conf.d as the following section. 
    ```gherkin=
    location / {
            # address run on local machine or VM
            proxy_pass http://127.0.0.1:8080/;
            # address run on docker environment 
            # proxy_pass http://app;
            proxy_http_version 1.1;
            proxy_set_header Host $http_host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    ```
3. go to /myapp1/goapp, then run the following command to copy DB_setting:
    ```
    cp DB_setting_example DB_setting.txt
    ```
   modify DB_setting.txt for logging in your DBs.
    ```gherkin=
    neo4j
    host:neo4j://localhost:7687
    username:{username}
    password:{password}

    postgresql
    host=localhost port=5432 dbname={dbname} user={username} password={password}
    host:localhost
    port:5432
    user:{username}
    password:{password}

    redis
    host:localhost:6379
    password:
    ```
    > replace {variable} the value you set.
4. Make sure postgres, neo4j, redis are running on your environment, then go to myapp1/goapp/ and run run.sh. The web-app will run and listen to :8080 port.
    ```
    sh run.sh
    ```


If you want to run this app on Docker envirnment.
1. Install docker and docker-compose if they have not been installed.
2. go to /myapp1/goapp, then run the following command to copy DB_setting:
    ```
    cp DB_setting_example DB_setting.txt
    ```
   modify DB_setting.txt for logging in your DBs.
    ```gherkin=
    neo4j
    host:neo4j://neo4j:7687
    username:{username}
    password:{password}

    postgresql
    host=postgres port=5432 dbname={dbname} user={username} password={password}
    host:localhost
    port:5432
    user:{username}
    password:{password}

    redis
    host:redis:6379
    password:
    ```
    > replace {variable} the value you set.
    
    Because of docker network (default mode: bridge), we must modify host value of each DB by their own hostname. Can see more about docker network bridge mode at: [Docker:Use bridge network](https://docs.docker.com/network/bridge/), [Docker: Networking with standalone containers](https://docs.docker.com/network/network-tutorial-standalone/).
3. go to /myapp1. And run the following command:
    ```
    sudo docker compose up
    ```
4. The web app will run and listen to :8080 port. And the nginx web server reverse proxy of the web app will run and listen to :80 port.

Web App architecture
---
![](https://i.imgur.com/QvnkgoC.png)


## Appendix and FAQ


