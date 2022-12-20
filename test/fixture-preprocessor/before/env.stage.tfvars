vpc_config = {
  name             = "vpc-stage"
  cidr_block       = "10.2.0.0/16"
  num_nat_gateways = 1
}

mysql_config = {
  name                              = "mysql-stage"
  version                           = "8.0"
  instance_class                    = "db.t3.micro"
  db_credentials_secrets_manager_id = "mysql-credentials"
}

frontend_config = {
  name                 = "frontend-stage"
  replicas             = 2
  docker_image_version = "v3.0.0"
}

backend_config = {
  name                 = "backend-stage"
  replicas             = 2
  docker_image_version = "v1.3.4"
}