variable region {
 default = "eu-central-1"
}

variable table_name {
 default = "sales-table"
}

provider aws {
 region = var.region
}
 
data aws_caller_identity current {}
 
locals {
 prefix = "go-scrape"
 account_id          = data.aws_caller_identity.current.account_id
 ecr_repository_name = "${local.prefix}-lambda-container"
 ecr_image_tag       = "latest"
}

resource aws_dynamodb_table dynamodb-table {
  name           = var.table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "url"

  attribute {
    name = "url"
    type = "S"
  }
}

resource aws_ecr_repository repo {
 name = local.ecr_repository_name
}
 
resource null_resource ecr_image {
 provisioner "local-exec" {
   command = <<EOF
           aws ecr get-login-password --region ${var.region} | docker login --username AWS --password-stdin ${local.account_id}.dkr.ecr.${var.region}.amazonaws.com
           docker build -t ${aws_ecr_repository.repo.repository_url}:${local.ecr_image_tag} --platform=linux/amd64 .
           docker push ${aws_ecr_repository.repo.repository_url}:${local.ecr_image_tag}
       EOF
 }
}
 
data aws_ecr_image lambda_image {
 depends_on = [
   null_resource.ecr_image
 ]
 repository_name = local.ecr_repository_name
 image_tag       = local.ecr_image_tag
}
 
resource aws_iam_role lambda {
 name = "${local.prefix}-lambda-role"
 assume_role_policy = <<POLICY
{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Action": "sts:AssumeRole",
           "Principal": {
               "Service": "lambda.amazonaws.com"
           },
           "Effect": "Allow"
       }
   ]
}
 POLICY
}

resource aws_iam_role_policy_attachment lambda-basic-exec-role {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource aws_iam_policy lambda-logs {
  name = "${local.prefix}-lambda-logs"
  path = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource aws_iam_role_policy_attachment lambda-logs-attachment {
  role = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.lambda-logs.arn
}

resource aws_iam_policy lambda-dynamo {
  name = "${local.prefix}-dynamo"
  path = "/"
  description = "IAM policy for DynamoDB access from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1582485790003",
      "Action": [
        "dynamodb:PutItem"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:dynamodb:*:*:*"
    }
  ]
}
EOF
}

resource aws_iam_role_policy_attachment lambda-dynamo {
  role = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.lambda-dynamo.arn
}

resource aws_lambda_function lambda-func {
 depends_on = [
  null_resource.ecr_image,
  aws_iam_role.lambda 
 ]
 function_name = "${local.prefix}-lambda"
 role = aws_iam_role.lambda.arn
 timeout = 300
 image_uri = "${aws_ecr_repository.repo.repository_url}@${data.aws_ecr_image.lambda_image.id}"
 package_type = "Image"
 environment {
    variables = {
      DYNAMODB_TABLE = aws_dynamodb_table.dynamodb-table.name,
      DYNAMODB_REGION = var.region
    }
  }
}
 
# we want to run this everyday at 4:20am 
resource aws_cloudwatch_event_rule lambda-cron {
  name = "${local.prefix}-cron"
  schedule_expression = "cron(10 0 * * ? *)"
}

resource aws_cloudwatch_event_target event-target {
  target_id = "runLambda"
  rule      = aws_cloudwatch_event_rule.lambda-cron.name
  arn       = aws_lambda_function.lambda-func.arn
}

resource aws_lambda_permission cloudwatch {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda-func.arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.lambda-cron.arn
}

output "lambda_name" {
 value = aws_lambda_function.lambda-func.id
}
