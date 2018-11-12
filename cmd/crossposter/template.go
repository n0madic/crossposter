package main

const indexTpl = `<!DOCTYPE html>
<html>
<head>
    <title>Crossposter</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/latest/css/bootstrap.min.css">
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/latest/css/bootstrap-theme.min.css">
</head>
<body>
<div class="container">
    <div class="jumbotron">
        <h2>Crossposter
            <small>by Nomadic</small>
        </h2>
        <p class="lead">Routing posts between different sources</p>
    </div>
    <div class="card bg-light">
        <h5 class="card-header"> Available routes </h5>
        <div class="card-body">
            <table class="table table-hover">
              <thead class="thead-light">
                <tr>
                    <th>Sources</th>
                    <th>Destinations</th>
                </tr>
              </thead>
              {{range $key, $value := .Sources}}
                <tr>
                    <td>[{{ $value.Entity }}] {{ $value.Description }} ({{ $key }})<br></td>
                    <td>{{ range $dest := $value.Destinations }}
                            [{{ (index $.Entities $dest).Type }}] {{ (index $.Entities $dest).Description }} ({{ $dest }})<br>
                    {{end}}</td>
                </tr>
              {{end}}
            </table>
        </div> <!-- card-body -->
    </div> <!-- card -->
    <footer>
        <div style="text-align: center;"><p><a href="https://github.com/n0madic/crossposter">GitHub</a> &copy; Nomadic 2018</p></div>
    </footer>
</div> <!-- container -->
</body>
</html>`
