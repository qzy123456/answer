<!DOCTYPE "html">
<head>
    <meta charset="utf-8">
    <title>五子棋对战</title>
    <script type="text/javascript" src="/public/javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="/public/javascripts/jquery.cookie.js"></script>
    <script src="/public/javascripts/bootstrap.min.js"></script>
    <script type="text/javascript" src="/public/javascripts/thinkbox/jquery.ThinkBox.min.js"></script>
    <link rel="stylesheet" type="text/css" href="/public/stylesheets/style.css"/>
    <link rel="stylesheet" type="text/css" href="/public/stylesheets/bootstrap.css"/>
    <link rel="stylesheet" type="text/css" href="/public/javascripts/thinkbox/thinkbox.css"/>

</head>
<style>
    canvas{
        display: block;
        margin: 50px auto;
        box-shadow: -2px -2px 2px #EFEFEF,5px 5px 5px #B9B9B9;
    }
</style>
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
        <div style="text-align: center"><h5 id="shows" style="color: red;">等待匹配对手...</h5></div>
    </div>
    <legend></legend>
    <div id="examMain">
        <canvas id="chess" width="450px" height="450px"></canvas>
    </div>
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
    var isAct;
    var chessBox = [];
    var chess = document.getElementById('chess');
    var context = chess.getContext('2d');
    context.strokeStyle = "#BFBFBF";
    var logo= new Image();
    var gameStatus = false,
        GameTimeStop = false,
        TkBox;
    logo.src = "public/images/木头.jpeg";
    logo.onload = function(){
        context.drawImage(logo,0,0,450,450);
        drawChessBoard();
    };
    function drawChessBoard(){
        for(var i=0;i<15;i++){
            chessBox[i]=[];
            for(var j=0;j<15;j++){
                chessBox[i][j]=0;
            }
        }
        for(var i=0;i<15;i++){
            context.moveTo(15+i*30,15);
            context.lineTo(15+i*30,435);
            context.moveTo(15,15+i*30);
            context.lineTo(435,15+i*30);
            context.stroke();
        }

    }
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
        //开始socket
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
                        $("#users .thumbnails").append('<li class="text-center" id="user_{0}" style="{4}"><img class="thumbnail" data-src="/public/images/header-{5}.jpeg" alt="{1}" src="/public/images/header-{5}.jpeg"><span>{2}</span><div class="gameStatus"></div><div id="lamp_{3}" class="lamp"></li>'.format(users[i].UserId, users[i].UserName, users[i].UserName, users[i].UserId, users[i].UserId != userId ? 'float:right; margin-right:20px;' : '',users[i].HeaderIndex));

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
                    chessBox = [];
                    var logo= new Image();
                    logo.src = "public/images/木头.jpeg";
                    logo.onload = function(){
                        context.drawImage(logo,0,0,450,450);
                        drawChessBoard();
                    };
                }

                ////开始游戏
                function onPlayGame(data) {
                    console.log("服务器返回数据",data.Params)
                    TkBox && TkBox.hide();

                    gameStatus = true;
                    $(".gameStatus").html('游戏中');

                    $("#examMain").show();

                    showTime(data.Params.GameTime);
                    isAct = getIsAct(data.Params.Users);
                    //该此用户下
                    console.log("当前用户信息",isAct);
                    if(isAct.IsAct && isAct.UserId == userId){
                        $("#shows").html("该你落子!")
                    }else {
                        $("#shows").html("等待对方落子!")
                    }
                    //Hand: 1
                    //IsAct: true
                   // Status: true
                   // UserId: 1655106139
                    chess.onclick = function(e){
                        if(isAct === false){
                            console.log("不改你下")
                            return;
                        }
                        var x = e.offsetX;
                        var y = e.offsetY;
                        var i = Math.floor(x/30); //i,j为索引序列号
                        var j = Math.floor(y/30);
                        if(chessBox[i][j]==0){
                            oneStep(i,j,isAct.Hand);
                            chessBox[i][j]=isAct.Hand;
                            console.log("棋盘",chessBox)
                            socket.send('{"Action":"Submit","UserId":' + userId + ',"Params":{"X": ' + i + ',"Y":' + j + '}}');
                        }
                    }
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

                //服务器下发另外一个用户
                function onGameResult(data) {
                    //{"Action":"GameResult","UserId":1655107294,"Params":{"Hand":1,"UserId":1655107294,"X":4,"Y":4},"Time":1655107334}
                    var x = data.Params.X;
                    var y = data.Params.Y;
                    oneStep(x,y,data.Params.Hand);
                    chessBox[x][y]=data.Params.Hand;
                }

                //退出游戏
                function onGameOver(data) {
                    TkBox && TkBox.hide();
                    GameTimeStop = true;
                    // {"Action":"GameOver","UserId":1655110358,"Params":{"OverUser":1655110353,"OverUserName":"1111"},"Time":1655110581}
                     $("#user_" + data.Params.OverUser).remove();
                    $("#shows").html("等待匹配对手...");
                    $.ThinkBox.success('用户' + data.Params.OverUserName + '退出了房间', {'modal': false});
                    $("#user_" + userId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');
                    $("#examMain").hide();
                    chessBox = [];
                    var logo= new Image();
                    logo.src = "public/images/木头.jpeg";
                    logo.onload = function(){
                        context.drawImage(logo,0,0,450,450);
                        drawChessBoard();
                    };
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
                    //{"Action":"EndGame","UserId":1655110358,
                    // "Params":{"Users":[{"UserId":1655110358,"Hand":2,"Status":true,"IsAct":false}],"W5110353},
                    // "Time":1655110581}

                    var users = data.Params.Users;
                    var Winner = data.Params.Winner;
                    gameStatus = false;
                    GameTimeStop = true;
                    $(".gameStatus").html('尚未准备');
                    $("#shows").html("等待匹配对手...")
                    $("#user_" + userId + " .gameStatus").html('<a href="javascript:ready();">开始游戏</a>');

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
                    var logo= new Image();
                    logo.src = "public/images/木头.jpeg";
                    logo.onload = function(){
                        context.drawImage(logo,0,0,450,450);
                        drawChessBoard();
                    };
                }

                function gameEndDialog(str) {
                    $.ThinkBox.confirm(str, {
                        "ok": function () {
                            this.hide();
                            $("#timer").hide();
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
                    console.log("ready",data.Params)
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


    var oneStep = function(i,j,me){
        context.beginPath();
        context.arc(15+i*30,15+j*30,13,0,2*Math.PI);
        context.closePath();
        var gradient = context.createRadialGradient(15+i*30,15+j*30,13,15+i*30,15+j*30,0);
        if(me === 1){
            gradient.addColorStop(0,"#0A0A0A");
            gradient.addColorStop(1,"#636766");
        }else{
            gradient.addColorStop(0,"#D1D1D1");
            gradient.addColorStop(1,"#F9F9F9");
        }

        context.fillStyle = gradient;
        context.fill();
    };


    // var computerAI = function(){
    //     var myScore = [];
    //     var computerScore = [];
    //     var max = 0; //保存最高分数；
    //     var u = 0, v =0; //保存坐标
    //     for(var i=0;i<15;i++){
    //         myScore[i] = [];
    //         computerScore [i] = [];
    //         for(var j=0;j<15;j++){
    //             myScore[i][j] = 0;
    //             computerScore[i][j] = 0;
    //         }
    //     }
    //     for (var i=0; i<15;i++) {
    //         for (var j=0;j<15;j++) {
    //             if(chessBox[i][j] == 0){
    //                 for(var k =0 ;k<count;k++){
    //                     if(wins[i][j][k]){
    //                         if(myWin[k]==1){
    //                             myScore[i][j]+= 200;
    //                         }else if(myWin[k]==2){
    //                             myScore[i][j]+= 400;
    //                         }else if(myWin[k]==3){
    //                             myScore[i][j]+= 2000;
    //                         }else if(myWin[k]==4){
    //                             myScore[i][j]+= 10000;
    //                         }
    //                         if(computerWin[k]==1){
    //                             computerScore[i][j]+= 220;
    //                         }else if(computerWin[k]==2){
    //                             computerScore[i][j]+= 420;
    //                         }else if(computerWin[k]==3){
    //                             computerScore[i][j]+= 2020;
    //                         }else if(computerWin[k]==4){
    //                             computerScore[i][j]+= 10020;
    //                         }
    //                     }
    //                 }
    //                 if(myScore[i][j]>max){
    //                     max = myScore[i][j];
    //                     u = i;
    //                     v = j;
    //                 }else if(myScore[i][j] == max){
    //                     if(computerScore[i][j] > computerScore[u][v]){
    //                         u = i;
    //                         v = j;
    //                     }
    //                 }
    //                 if(computerScore[i][j]>max){
    //                     max = computerScore[i][j];
    //                     u = i;
    //                     v = j;
    //                 }else if(computerScore[i][j] == max){
    //                     if(myScore[i][j] > myScore[u][v]){
    //                         u = i;
    //                         v = j;
    //                     }
    //                 }
    //             }
    //         }
    //     }
    //     oneStep(u,v,false);
    //     chessBox[u][v] = 2;
    //     for(var k=0;k < count; k++){
    //         if(wins[u][v][k]) {
    //             computerWin[k]++;
    //             myWin[k] = 6; //设置异常值
    //             if(computerWin[k] == 5) {
    //                 window.alert("计算机赢了");
    //                 over = true;
    //             }
    //         }
    //     }
    //     if(!over){
    //         me=!me;
    //     }
    // }
    function getIsAct(users) {
        for(var i = 0; i < users.length; i++) {
            if(users[i].IsAct == true && users[i].UserId == userId) {
                return users[i];
            }
        }
        return false;
    }
</script>