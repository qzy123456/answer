<!DOCTYPE html>
<html>
<head>
    <title>用户登陆</title>
    <link href="/public/stylesheets/bootstrap.min.css" rel="stylesheet" media="screen">
</head>
<body screen_capture_injected="true">
<div class="container-fluid">
    <form class="form-horizontal" method="post" action="/">
        <fieldset>
            <legend>请输入昵称</legend>
            <div class="form-actions" style="text-align: center">
                <label class="control-label" for="user">用户名</label>
                <div class="controls">
                    <input type="text" class="input-xlarge" id="userName" name="user" required="true" />
                </div>
            </div>
            <div class="form-actions" style="text-align: center">
                <button type="submit" class="btn btn-primary">登陆</button>
            </div>
        </fieldset>
    </form>
</div>
<script src="/public/javascripts/jquery-1.7.2.min.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>
</body>
</html>