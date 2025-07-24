resource "doppler_integration_aws_iam_user_keys" "i_aws_iam_uk" {
  name            = "TF AWS IAM UK"
  assume_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/xxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_aws_iam_user_keys" "rs_aws_iam_uk" {
  integration         = doppler_integration_aws_iam_user_keys.i_aws_iam_uk.id
  project             = "backend"
  config              = "dev"
  name                = "AWS_IAM_UK"
  rotation_period_sec = 2592000
  username            = "xxxxxxxxxxxxx"
}
