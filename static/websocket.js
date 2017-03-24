function AttachProxy(id, commandId) {
    var shell = $("#" + id);

    shell.WriteLine = function(line) {
        var newP;
        if (line[0] == '#') {
            newP = shell.append("<p class='comment'>" + line + "</p>");
        }
        else {
            newP = shell.append("<p>" + line + "</p>");
        }
        shell.scrollTop(shell.scrollTop() + newP.position().top + newP.height());
    }

    var socket = new WebSocket("ws://localhost:8080/socket");
    socket.onopen = function(event) {
        shell.WriteLine("# Connected to remote websocket!");
    }
    socket.onerror = function(event) {
        shell.WriteLine("# Error: " + event)
    }
    socket.onmessage = function(event) {
        shell.WriteLine(event.data);
    }
    socket.onclose = function(event) {
        shell.WriteLine("# Remote connection lost! " + event.code)
    }

    var command = $("#" + commandId);
    command.keyup(function(event) {
        if (event.keyCode == 13) {
            var cmd = command.val();
            command.val("");

            shell.WriteLine(cmd);
            if (cmd[0] != '#') {
                socket.send(cmd + "\n");
            }
        }
    });
}
