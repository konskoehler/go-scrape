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

resource aws_iam_policy go-scrape-lambda_logging {
  name = "go-eat-lambda_logging"
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

resource aws_iam_role_policy_attachment go-scrape-lambda_logs {
  role = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.go-scrape-lambda_logging.arn
}

resource aws_iam_policy go-scrape-dynamo {
  name = "go-scrape-dynamo"
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

resource aws_iam_role_policy_attachment go-scrape-dynamo {
  role = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.go-scrape-dynamo.arn
}

resource aws_lambda_function go-scrape {
 depends_on = [
  null_resource.ecr_image,
  aws_iam_role.lambda 
 ]
 function_name = "${local.prefix}-lambda"
 role = aws_iam_role.lambda.arn
 timeout = 30
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
resource aws_cloudwatch_event_rule go-scrape-cron {
  name                = "go-scrape-cron"
  schedule_expression = "cron(20 4 * * ? *)"
}

resource aws_cloudwatch_event_target go-scrape-lambda {
  target_id = "runLambda"
  rule      = aws_cloudwatch_event_rule.go-scrape-cron.name
  arn       = aws_lambda_function.go-scrape.arn
}

resource aws_lambda_permission go-scrape-cloudwatch {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.go-scrape.arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.go-scrape-cron.arn
}

output "lambda_name" {
 value = aws_lambda_function.go-scrape.id
}
