<!DOCTYPE html>
<html>
<head>
    <title>请起一个昵称</title>
    <link href="/public/stylesheets/bootstrap.min.css" rel="stylesheet" media="screen">
</head>
<body screen_capture_injected="true">
<div class="container-fluid">
    <form class="form-horizontal" method="post" action="/game/" onsubmit="return verify()">
        <fieldset>
            <legend>用户昵称</legend>
            <div class="form-actions">
                <label class="control-label" for="user">昵称</label>
                <div class="controls">
                    <input type="text" class="input-xlarge" id="user" name="user" required="true" />
                </div>
            </div>
            <div class="form-actions">
                <button type="submit" class="btn btn-primary">登陆</button>
            </div>
        </fieldset>
    </form>
</div>
<script src="/public/javascripts/jquery-1.7.2.min.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>
</body>
</html>
<script type="text/javascript">
    function verify() {
        var name = $("#user").val();
        name = name.trim();
        if (name.length <= 0){
            alert("名字不能为空哦!");
            return false;
        }
    }
</script>