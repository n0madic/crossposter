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
        <h5 class="card-header"> Available pipelines </h5>
        <div class="card-body">
            <table class="table table-hover">
              <thead class="thead-light">
                <tr>
                    <th>Role</th>
                    <th>Entity</th>
                    <th>Sources/Destinations</th>
                </tr>
              </thead>
              {{- range $entity := .Entities}}
                <tr>
                <td>{{ $entity.Role }}</td>
                <td>[{{ $entity.Type }}] {{ $entity.Description }}</td>
                {{- if eq $entity.Role "producer" -}}
                <td>{{ StringsJoin $entity.Sources "<br>" }}</td>
                {{- end }}
                {{- if eq $entity.Role "consumer" -}}
                <td>{{ StringsJoin $entity.Destinations "<br>" }}</td>
                {{- end }}
                </tr>
              {{- end }}
            </table>
        </div> <!-- card-body -->
    </div> <!-- card -->
    <footer>
        <div style="text-align: center;"><p><a href="https://github.com/n0madic/crossposter">GitHub</a> &copy; Nomadic 2018</p></div>
    </footer>
</div> <!-- container -->
</body>
</html>`
