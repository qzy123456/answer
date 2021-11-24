<html>
<head>
<meta charset="utf-8">
<title>双人答题对战</title>
<script type="text/javascript" src="/public/javascripts/jquery.min.js"></script>
<script type="text/javascript" src="/public/javascripts/jquery.cookie.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>
<script type="text/javascript" src="/public/javascripts/thinkbox/jquery.ThinkBox.min.js"></script>
<link rel="stylesheet" type="text/css" href="/public/stylesheets/style.css"  />
<link rel="stylesheet" type="text/css" href="/public/stylesheets/bootstrap.css"  />
<link rel="stylesheet" type="text/css" href="/public/javascripts/thinkbox/thinkbox.css"  />

</head>
<body>
<button id="outRoom" type="button" class="btn btn-warning">退出房间</button>
<input type="hidden" id="userName" value="{{ .userName}}">
<input type="hidden" id="userId" value="{{ .userId}}">
<div id="main">
    <div class="a">
        <div id="users">  <!-- 用户列表 -->
        <div class="row-fluid">
            <ul class="thumbnails">
            </ul>
          </div>
        </div>
        <div id="timer">50</div>
    </div>
    <legend></legend>
    <div id="examMain">
        <div id="examTitle" class="text-center"></div>
        <div id="examOption"></div>
    </div>   <!-- 答题处理 -->
</div>

</body>
</html>
<script type="text/javascript">
    String.prototype.format = function() {
        var args = arguments;
        return this.replace(/\{(\d+)\}/g, function(m, i) { return args[i]; });
    };
    var userName = $("#userName").val();
    var userId = $("#userId").val();
    var host = "127.0.0.1:9999";
    $(document).ready(function() {
        var socket = new WebSocket("ws://" + host + "/ws?userId="+userId);
        socket.onopen = function(evt) {
            console.log("socket连接成功");
            socket.send('{"Action":"Login","UserId":' + userId + ',"UserName":"'+userName+'","Params":{}}');  //发送登录socket连接
        }
        socket.onclose = function(evt) {
            onDisconnect();
        }
        socket.onmessage = function(evt) {
            var data = eval("("+evt.data+")");  //解析JSON数据
            console.log("接收到Socket信息：");
            console.log(data)
            switch(data.Action) {
                case "offline":
                    onOnline(data);
                    break;
                case "JoinRoom":
                    onJoinRoom(data);
                    break;
                case "PlayGame":
                    onPlayGame(data);
                    break;
                case "EndGame":
                    onEndGame(data);
                    break;
                case "GameResult":
                    onGameResult(data);
                    break;
                case "GameOver":
                    onGameOver(data);
                    break;
                case "OutRoom":
                    onOutRoom(data);
                    break;
                case "Ready":
                    onReady(data);
                    break;
            }
        }

        var gameStatus = false,
            GameTimeStop = false,
            TkBox;

        function ready() { //准备
            $("#user_" + userId + ' .gameStatus').html("准备中");
             socket.send('{"Action":"Ready","UserId":' + userId + ',"Params":{}}');

        }

        $("#outRoom").click(function () { //退出房间
            socket.send('{"Action":"OutRoom","UserId":' + userId + ',"Params":{}}');
        });

        ////加入房间socket接收
        function onJoinRoom(data) {
            TkBox && TkBox.hide();
            $("#users .thumbnails").html("");
            var users = data.Params.Users;
            var otherUser;
            for(var i=0; i < users.length; i++) {
                $("#users .thumbnails").append('<li class="text-center" id="user_{0}" style="{4}"><img class="thumbnail" data-src="/public/images/header.jpg" alt="{1}" src="/public/images/header.jpg">{2}<div class="gameStatus"></div><div id="lamp_{3}" class="lamp"></li>'.format(users[i].UserId, users[i].UserName, users[i].UserName, users[i].UserId, users[i].UserId != userId ? 'float:right; margin-right:20px;' : ''));

                if( users[i].UserId != userId ) {
                    otherUser = users[i];
                }
            }

            var rmUsers = data.Params.Room.Users;
            for(var i=0; i < rmUsers.length; i++) {
                var str = rmUsers.length >= 2 ? '准备中' : (rmUsers[i].UserId != userId ? '尚未准备' : '<a href="javascript:ready();">开始游戏</a>');
                $("#user_" + rmUsers[i].UserId + " .gameStatus").html(str);
            }

            var str;
            console.log("另一个人",otherUser);
            if( data.UserId == userId ) {
                str = "您加入了房间";
            } else if( otherUser ) {
                str = "玩家" + otherUser.UserName + "加入了房间";
            }
            str && $.ThinkBox.success(str, {'modal': false});
        }

        ////开始游戏
        function onPlayGame(data) {
            TkBox && TkBox.hide();

            gameStatus = true;
            $(".gameStatus").html('游戏中');

            for(var i = 0; i < data.Params.Users.length; i++) {
                $("#lamp_" + data.Params.Users[i].UserId).html("");
                for( var j = 0; j < data.Params.Users[i].Views; j++ ) {
                    $("#lamp_" + data.Params.Users[i].UserId).append("<i></i>");
                }
            }
            $("#examMain").show();
            $("#examTitle").html("<h1>{0}</h1>".format(data.Params.Exam.ExamQuestion));
            $("#examOption").html("");  //清空
            var isAct = true;
            for(var i = 0; i < data.Params.Exam.ExamOption.length; i++) {
                $("#examOption").append('<button aid="{0}" type="button" class="btn btn-large exmp {1}">{2}</button>'.format(i, isAct ? 'submit' : 'disabled', data.Params.Exam.ExamOption[i]));
            }
            showTime(data.Params.GameTime);

            $(".submit").bind("click", function () { //提交答案
                TkBox = $.ThinkBox.loading('正在提交答案...');
                $("#examOption button").removeClass("submit");
                $("#examOption buttom").attr("onclick", "");
                console.log("提交答案");
                socket.send('{"Action":"Submit","UserId":' + userId + ',"Params":{"AnswerId": '+$(this).attr("aid")+'}}');
            });
        }

        var duration, endTime;
        function showTime(GameTime) {
            $("#timer").show();
            GameTimeStop = false;
            duration = GameTime * 1000 - 100;
            endTime  = new Date().getTime() + duration + 100;
            interval();
        }

        function interval() {
            var n=(endTime-new Date().getTime())/1000;
            if(n<0 || GameTimeStop == true) return;
            document.getElementById("timer").innerHTML = n.toFixed(3);
            setTimeout(interval, 10);
        }
        //答案提交验证
        function onGameResult(data) {
            GameTimeStop = true
            TkBox && TkBox.hide();
            $("#examOption button[aid="+data.Params.Answer+"]").addClass("btn-success");
            if(data.Params.IsOk == true) {
                TkBox = $.ThinkBox.success(data.Params.UserId == userId ? '恭喜你答对了' : '答对了', {'delayClose':1000});
            } else {
                $("#examOption button[aid="+data.Params.UserAnswer+"]").addClass("btn-danger");
                TkBox = $.ThinkBox.error(data.Params.UserId == userId ? '对不起你答错了' : '答错了', {'delayClose':1000});
            }
        }
        //退出游戏
        function onGameOver(data) {
            TkBox && TkBox.hide();
            GameTimeStop = true;
            $("#user_" + data.Params.OverUser).remove();
            $.ThinkBox.success('用户' + data.Params.OverUserName + '退出了房间', {'modal': false});
            $("#user_" + UserId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');

            gameEndDialog('恭喜 ，你赢了！！ ^_^');
        }
        //退出房间
        function onOutRoom(data) {
            window.location.href="http://"+host;
        }

        //结束游戏
        function onEndGame(data) {
            console.log("end",data)
            TkBox && TkBox.hide();
            var users = data.Params.Users;
            gameStatus = false;
            GameTimeStop = true;
            $(".gameStatus").html('尚未准备');
            $("#user_" + userId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');

            for(var i = 0; i < users.length; i++) {
                if( users[i+1].Status == 1 ) {  //机器人自动准备
                    $("#user_" + users[i+1].UserId + " .gameStatus").html('准备中');
                }
                if( users[i].UserId == userId) {
                    if(users[i].UserId == userId)
                        $("#lamp_" + users[i+1].UserId).html("");
                    else
                        $("#lamp_" + users[i].UserId).html("");
                    gameEndDialog('恭喜 ，你赢了！！ ^_^');
                } else {
                    $("#lamp_" + userId).html("");
                    gameEndDialog('::>_<::，你输了');
                }
                break;
            }
        }

        function gameEndDialog(str) {
            $.ThinkBox.confirm(str, {
                "ok": function() {
                    this.hide();
                    $("#timer").hide();
                    $("#examMain").hide();
                    $(".lamp").html("");
                    ready();
                },
                "cancel": function() {
                    $("#outRoom").click();
                    this.hide();
                    $("#timer").hide();
                    $("#examMain").hide();
                    $(".lamp").html("");
                },
                "modal": true,
                "close": false,
                "okVal": "再来一局",
                "cancelVal": "退出房间"
            });
        }

        //接收准备消息
        function onReady(data) {
            console.log(data.Params.UserId)
            $("#user_" + data.Params.UserId + ' .gameStatus').html("准备中");
        }


        function onDisconnect() { //连接服务器失败
            console.log("连接服务器失败");
            $.ThinkBox.error('连接服务器失败', {'delayClose':2000});
        }
    });

</script>