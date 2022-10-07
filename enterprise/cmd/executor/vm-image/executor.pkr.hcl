packer {
    required_plugins {
        amazon = {
            version = ">= 1.0.0"
            source  = "github.com/hashicorp/amazon"
        }
        googlecompute = {
            version = ">= 1.0.0"
            source  = "github.com/hashicorp/googlecompute"
        }
    }
}

variable "name" {
    type = string
}

variable "version" {
    type = string
}

variable "srcCliVersion" {
    type = string
}

variable "awsAccessKey" {
    type = string
    sensitive = true
}

variable "awsSecretKey" {
    type = string
    sensitive = true
}

variable "awsMaxAttempts" {
    type = number
}

variable "awsPollDelaySeconds" {
    type = number
}

source "googlecompute" "gcp" {
    project_id = "sourcegraph-ci"
    source_image_project_id = "ubuntu-os-cloud"
    source_image_family = "ubuntu-2004-lts"
    disk_size = "10"
    ssh_username = "packer"
    zone = "us-central1-c"
    image_licenses = ["projects/vm-options/global/licenses/enable-vmx"]
    disk_type = "pd-ssd"
    image_name = var.name
    image_description = "Convenience image to run Sourcegraph executors. See github.com/sourcegraph/terraform-google-executors for how to use it."
    image_storage_locations = ["us"]
    tags = ["packer"]
    account_file = "builder-sa-key.json"
}

source "amazon-ebs" "aws" {
    ami_name = var.name
    ami_description = "Convenience image to run Sourcegraph executors. See github.com/sourcegraph/terraform-google-executors for how to use it."
    ssh_username = "ubuntu"
    instance_type = "t3.micro"
    source_ami_filter {
        filters {
          virtualization-type = "hvm"
          name = "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"
          root-device-type = "ebs"
        }
        owners = ["099720109477"]
        most_recent = true
      }
      region = "us-west-2"
      vpc_id = "vpc-0fae37a99a5156b91"
      subnet_id = "subnet-0a71d7cd03fea6317"
      associate_public_ip_address = true
      access_key = var.awsAccessKey
      secret_key = var.awsSecretKey
      aws_polling {
        delay_seconds = var.awsPollDelaySeconds
        max_attempts = var.awsMaxAttempts
      }
      shutdown_behavior = "terminate"
      ami_regions = ["us-west-1", "us-west-2", "us-east-1", "us-east-2", "eu-west-2"]
      tags = {
        Name = var.name
        OS_Version = "Ubuntu"
        Release = "Latest"
        Base_AMI_Name = "{{ .SourceAMIName }}"
        Extra = "{{ .SourceAMITags.TagName }}"
      }
}

build {
    sources = [
        "source.googlecompute.gcp",
        "source.amazon-ebs.aws"
    ]

    provisioner "file" {
        sources = ["executor"]
        destination = "/tmp/"
    }

    provisioner "file" {
        sources = ["executor-vm"]
        destination = "/tmp"
    }

    provisioner "shell" {
        execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo -E bash {{ .Path }}"
        script = "install.sh"
        override = {
            "gcp" = {
                environment_vars = [
                    "SRC_CLI_VERSION=${var.srcCliVersion}",
                    "VERSION=${var.version}",
                    "PLATFORM_TYPE=gcp"
                ]
            },
            "aws" = {
                environment_vars = [
                    "SRC_CLI_VERSION=${var.srcCliVersion}",
                    "VERSION=${var.version}",
                    "PLATFORM_TYPE=aws"
                ]
            }
        }
    }
}
