<!DOCTYPE "html">
<head>
    <meta charset="utf-8">
    <title>双人答题对战</title>
    <script type="text/javascript" src="/public/javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="/public/javascripts/jquery.cookie.js"></script>
    <script src="/public/javascripts/bootstrap.min.js"></script>
    <script type="text/javascript" src="/public/javascripts/thinkbox/jquery.ThinkBox.min.js"></script>
    <link rel="stylesheet" type="text/css" href="/public/stylesheets/style.css"/>
    <link rel="stylesheet" type="text/css" href="/public/stylesheets/bootstrap.css"/>
    <link rel="stylesheet" type="text/css" href="/public/javascripts/thinkbox/thinkbox.css"/>

</head>
<body>
<button id="outRoom" type="button" class="btn btn-warning">退出房间</button>
<input type="hidden" id="userName" value="">
<input type="hidden" id="userId" value="{{ .userId}}">
<input type="hidden" id="Host" value="{{ .Host}}">
<input type="hidden" id="questionId" value="">
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
        <div id="examImg" class="text-center"></div>
        <div id="examOption"></div>
    </div>   <!-- 答题处理 -->
</div>
<div class="modal fade send-pop" id="add-name" tabindex="-1" role="dialog"
     aria-labelledby="send-res">
    <div class="modal-dialog" role="document">
        <div class="modal-content">

                <div class="modal-header">
                    <h4 class="modal-title" id="myModalLabel">请起一个昵称</h4>
                </div>
                <div class="modal-body">
                    <div class="form-inline" style="margin-top: 15px;">
                        <label for="de-family-star">用户昵称: </label>
                        <input type="text" class="input-large"  style="margin-left: 10px;"
                               id="de-family-star" />
                    </div>
                </div>

                <div class="modal-footer">
                    <button id="submit-name" class="btn btn-primary btn-send-on" onclick="return addName()">确定</button>
                </div>

        </div>
    </div>
</div>

</body>
</html>
<script type="text/javascript">
    String.prototype.format = function () {
        var args = arguments;
        return this.replace(/\{(\d+)\}/g, function (m, i) {
            return args[i];
        });
    };
    var userName = $("#userName").val();
    var userId = $("#userId").val();
    var host = $("#Host").val();
    //起名字
    function addName() {
        $("#submit-family-star").attr({ "disabled": "disabled" });
        userName = $("#de-family-star").val();
        userName = userName.trim();
        if(userName.length < 1 || userName == ''){
            $.ThinkBox.error('名字不能为空', {'delayClose': 1000});
            return false
        }
        $("#userName").val(userName);
        $("#add-name").modal("hide");
        sockets();
    }

    function sockets(){
        $(document).ready(function () {
                var socket = new WebSocket("ws://" + host + "/ws?userId=" + userId);
                socket.onopen = function (evt) {
                    console.log("socket连接成功");
                    socket.send('{"Action":"Login","UserId":' + userId + ',"UserName":"' + userName + '","Params":{}}');  //发送登录socket连接
                }
                socket.onclose = function (evt) {
                    onDisconnect();
                }
                socket.onmessage = function (evt) {
                    var data = eval("(" + evt.data + ")");  //解析JSON数据
                    console.log("接收到Socket信息：");
                    console.log(data)
                    switch (data.Action) {
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
                    for (var i = 0; i < users.length; i++) {
                        $("#users .thumbnails").append('<li class="text-center" id="user_{0}" style="{4}"><img class="thumbnail" data-src="/public/images/header.jpg" alt="{1}" src="/public/images/header.jpg"><span>{2}</span><div class="gameStatus"></div><div id="lamp_{3}" class="lamp"></li>'.format(users[i].UserId, users[i].UserName, users[i].UserName, users[i].UserId, users[i].UserId != userId ? 'float:right; margin-right:20px;' : ''));

                        if (users[i].UserId != userId) {
                            otherUser = users[i];
                        }
                    }

                    var rmUsers = data.Params.Room.Users;
                    for (var i = 0; i < rmUsers.length; i++) {
                        var str = rmUsers[i].Status == 1 ? '准备中' : (rmUsers[i].UserId != userId ? '尚未准备' : '<a href="javascript:ready();">开始游戏</a>');
                        $("#user_" + rmUsers[i].UserId + " .gameStatus").html(str);
                    }

                    var str;
                    console.log("另一个人", otherUser);
                    if (data.UserId == userId) {
                        str = "您加入了房间";
                    } else if (otherUser) {
                        str = "玩家" + otherUser.UserName + "加入了房间";
                    }
                    str && $.ThinkBox.success(str, {'modal': false});
                }

                ////开始游戏
                function onPlayGame(data) {
                    TkBox && TkBox.hide();

                    gameStatus = true;
                    $(".gameStatus").html('游戏中');

                    $("#examMain").show();
                    $("#examTitle").html("<h1>{0}</h1>".format(data.Params.Exam.question_title));
                    $("#examOption").html("");  //清空
                    $("#examImg").html("");  //清空
                    if (data.Params.Exam.question_img.length > 0) {
                        $("#examImg").html("<img src={0}>".format(data.Params.Exam.question_img));
                    }
                    $("#questionId").val("");
                    var isAct = true;
                    var answers = data.Params.Exam.answers;
                    answers = answers.substring(1, answers.length - 1)
                    var option = answers.split(",");
                    $("#questionId").val(data.Params.Exam.id);
                    for (var i = 0; i < option.length; i++) {
                        $("#examOption").append('<button aid="{0}" type="button" class="btn btn-large exmp {1}">{2}</button>'.format(i, isAct ? 'submit' : 'disabled', option[i]));
                    }
                    showTime(data.Params.GameTime);

                    $(".submit").bind("click", function () { //提交答案
                        TkBox = $.ThinkBox.loading('正在提交答案...');
                        $("#examOption button").removeClass("submit");
                        $("#examOption buttom").attr("onclick", "");
                        console.log("提交答案");
                        //拼接答案
                        var questionId = $("#questionId").val();

                        socket.send('{"Action":"Submit","UserId":' + userId + ',"Params":{"AnswerId": ' + $(this).attr("aid") + ',"QuestionId":' + questionId + '}}');
                    });
                }

                var duration, endTime;

                function showTime(GameTime) {
                    $("#timer").show();
                    GameTimeStop = false;
                    duration = GameTime * 1000;
                    endTime = new Date().getTime() + duration;
                    interval();
                }

                function interval() {
                    var n = (endTime - new Date().getTime()) / 1000;
                    if (n < 0 || GameTimeStop == true) {
                        document.getElementById("timer").innerHTML = '0.00';
                        return;
                    }
                    document.getElementById("timer").innerHTML = n.toFixed(2);
                    setTimeout(interval, 10);
                }

                //答案提交验证
                function onGameResult(data) {
                    GameTimeStop = true
                    TkBox && TkBox.hide();
                    $("#examOption button[aid=" + data.Params.Answer + "]").addClass("btn-success");
                    if (data.Params.IsOk == true) {
                        TkBox = $.ThinkBox.success(data.Params.UserId == userId ? '恭喜你答对了' : '对方答对了', {'delayClose': 1000});
                    } else {
                        $("#examOption button[aid=" + data.Params.UserAnswer + "]").addClass("btn-danger");
                        TkBox = $.ThinkBox.error(data.Params.UserId == userId ? '对不起你答错了' : '对方答错了', {'delayClose': 1000});
                    }
                }

                //退出游戏
                function onGameOver(data) {
                    TkBox && TkBox.hide();
                    GameTimeStop = true;
                    $("#user_" + data.Params.OverUser).remove();
                    $.ThinkBox.success('用户' + data.Params.OverUserName + '退出了房间', {'modal': false});
                    $("#user_" + userId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');
                    $("#examMain").hide();
                    gameEndDialog('恭喜 ，你赢了！！ ^_^');
                }

                //退出房间
                function onOutRoom(data) {
                    TkBox && TkBox.hide();
                    if (data.Params.OverUser == userId) {
                        window.location.href = "http://" + host
                    } else {
                        $("#user_" + data.Params.OverUser).remove();

                        $.ThinkBox.success('用户' + data.Params.OverUserName + '退出了房间', {'modal': false});
                    }
                }

                //结束游戏
                function onEndGame(data) {
                    console.log("end", data)
                    TkBox && TkBox.hide();
                    var users = data.Params.Users;
                    var Winner = data.Params.Winner;
                    gameStatus = false;
                    GameTimeStop = true;
                    $(".gameStatus").html('尚未准备');
                    $("#user_" + userId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');
                    $("#examMain").hide();
                    for (var i = 0; i < users.length; i++) {
                        if (users[i + 1].Status == 1) {  //机器人自动准备
                            $("#user_" + users[i + 1].UserId + " .gameStatus").html('准备中');
                        }
                        if (Winner == userId) {
                            $("#lamp_" + users[i + 1].UserId).html("");
                            $("#lamp_" + users[i].UserId).html("");
                            gameEndDialog('恭喜 ，你赢了！！ ^_^');
                        } else if (Winner == 0) {
                            $("#lamp_" + users[i + 1].UserId).html("");
                            $("#lamp_" + users[i].UserId).html("");
                            gameEndDialog('平局！！ ^_^');
                        } else {
                            $("#lamp_" + userId).html("");
                            gameEndDialog('::>_<::，你输了');
                        }
                        break;
                    }
                }

                function gameEndDialog(str) {
                    $.ThinkBox.confirm(str, {
                        "ok": function () {
                            this.hide();
                            $("#timer").hide();
                            $("#examMain").hide();
                            $(".lamp").html("");
                            ready();
                        },
                        "cancel": function () {
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
                    console.log(data.Params)
                    $("#user_" + data.Params.UserId + ' .gameStatus').html("准备中");
                }


                function onDisconnect() { //连接服务器失败
                    console.log("连接服务器失败");
                    $.ThinkBox.error('连接服务器失败', {'delayClose': 2000});
                }

        });
    }
    $(document).ready(function () {
        //没起名字，主动
        if (userName.length < 1 || userName == '') {
            $("#add-name").modal("show");
            return false
        }

    });

</script>