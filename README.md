# export_param_store
Fetch parameters from AWS Systems Manager Parameter Store 
and store them into environment variables.

## Example Usage
```
# $(export_param_store --region ap-northeast-1 --env prod --identifier service_name MYSQL_USER_NAME MYSQL_USER_PASSWORD)
```

By this command, `MYSQL_USER_NAME` and `MYSQL_USER_PASSWORD` are exported as environment variables.

* `MYSQL_USER_NAME` corresponds a parameter named `prod.service_name.mysql_user_name` in Parameter Store.
* `MYSQL_USER_PASSWORD` corresponds a parameter named `prod.service_name.mysql_user_password` in Parameter Store.

### How does this work?
This command fetches parameters from Parameter Store
and print them like as `export (ENV_VAR_NAME)=(env_var_value)`.

The parameters should be named `(env).(identifier).(environment_var_name)`.

If we have two parameters in Parameter Store,

* `prod.service_name.mysql_user_name`
* `prod.service_name.mysql_user_password`

this command will print values like as following.

```
# export_param_store --region ap-northeast-1 --env prod --identifier service_name MYSQL_USER_NAME MYSQL_USER_PASSWORD

export MYSQL_USER_NAME=0gajun
export MYSQL_USER_PASSWORD=hogehoge
```

So, what only you have to do is evaluating this output using `$()`.

Yeah! We can export environment variables from Parameter Store much easily!

## Installation

```
# git clone https://github.com/0gajun/export_param_store.git
# glide install
# go install
```

## Author
0gajun <oga.ivc.s27@gmail.com>
