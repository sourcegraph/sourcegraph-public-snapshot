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
        amazon-ami-management = {
            version = ">= 1.3.0"
            source  = "github.com/wata727/amazon-ami-management"
        }
    }
}

variable "image_family" {
    type = string
}

variable "tagged_release" {
    type = bool
}

variable "name" {
    type = string
}

variable "version" {
    type = string
}

variable "src_cli_version" {
    type = string
}

variable "aws_access_key" {
    type      = string
    sensitive = true
}

variable "aws_secret_key" {
    type      = string
    sensitive = true
}

variable "aws_max_attempts" {
    type = number
}

variable "aws_poll_delay_seconds" {
    type = number
}

variable "aws_regions" {
    type = list(string)

    validation {
        condition     = length(var.aws_regions) > 0
        error_message = "Must set at least 1 AWS region."
    }
}

locals {
    aws_ami_management_identifier = var.tagged_release ? {} : { Amazon_AMI_Management_Identifier = var.image_family }
}

source "googlecompute" "gcp" {
    project_id              = "sourcegraph-ci"
    source_image_project_id = ["ubuntu-os-cloud"]
    source_image_family     = "ubuntu-2004-lts"
    disk_size               = 10
    ssh_username            = "packer"
    zone                    = "us-central1-c"
    image_licenses          = ["projects/vm-options/global/licenses/enable-vmx"]
    disk_type               = "pd-ssd"
    image_name              = var.name
    image_description       = "Convenience image to run Sourcegraph executors. See github.com/sourcegraph/terraform-google-executors for how to use it."
    image_storage_locations = ["us"]
    tags                    = ["packer"]
    account_file            = "builder-sa-key.json"
}

source "amazon-ebs" "aws" {
    ami_name        = var.name
    ami_description = "Convenience image to run Sourcegraph executors. See github.com/sourcegraph/terraform-aws-executors for how to use it."
    ssh_username    = "ubuntu"
    instance_type   = "t3.micro"

    source_ami_filter {
        filters = {
            virtualization-type = "hvm"
            name                = "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"
            root-device-type    = "ebs"
        }
        owners      = ["099720109477"]
        most_recent = true
    }

    region                      = "us-west-2"
    vpc_id                      = "vpc-0fae37a99a5156b91"
    subnet_id                   = "subnet-0a71d7cd03fea6317"
    associate_public_ip_address = true
    access_key                  = var.aws_access_key
    secret_key                  = var.aws_secret_key

    aws_polling {
        delay_seconds = var.aws_poll_delay_seconds
        max_attempts  = var.aws_max_attempts
    }

    shutdown_behavior = "terminate"
    ami_regions       = var.aws_regions
    tags              = merge({
        Name          = var.name
        OS_Version    = "Ubuntu"
        Release       = "Latest"
        Base_AMI_Name = "{{ .SourceAMIName }}"
        Extra         = "{{ .SourceAMITags.TagName }}"
    }, local.aws_ami_management_identifier)
}

build {
    sources = [
        "source.googlecompute.gcp",
        "source.amazon-ebs.aws"
    ]

    provisioner "file" {
        sources     = ["executor"]
        destination = "/tmp/"
    }

    provisioner "file" {
        sources     = ["executor-vm.tar"]
        destination = "/tmp/executor-vm.tar"
    }

    provisioner "shell" {
        execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo -E bash {{ .Path }}"
        script          = "install.sh"
        override        = {
            "gcp" = {
                environment_vars = [
                    "SRC_CLI_VERSION=${var.src_cli_version}",
                    "VERSION=${var.version}",
                    "PLATFORM_TYPE=gcp"
                ]
            },
            "aws" = {
                environment_vars = [
                    "SRC_CLI_VERSION=${var.src_cli_version}",
                    "VERSION=${var.version}",
                    "PLATFORM_TYPE=aws"
                ]
            }
        }
    }

    post-processor "amazon-ami-management" {
        only       = ["amazon-ebs.aws"]
        access_key = var.aws_access_key
        secret_key = var.aws_secret_key
        regions    = ["us-west-2"]
        identifier = var.image_family
        keep_days  = 60
    }
}
