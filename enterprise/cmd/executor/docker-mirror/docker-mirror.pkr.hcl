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
        azure = {
            version = ">= 1.4.0"
            source  = "github.com/hashicorp/azure"
        }
    }
}

# variable "name" {
#     type = string
# }
#
# variable "aws_access_key" {
#     type = string
#     sensitive = true
# }
#
# variable "aws_secret_key" {
#     type = string
#     sensitive = true
# }
#
# variable "aws_max_attempts" {
#     type = number
# }
#
# variable "aws_poll_delay_seconds" {
#     type = number
# }
#
# variable "aws_regions" {
#     type = list(string)
#
#     validation {
#         condition = length(var.aws_regions) > 0
#         error_message = "Must set at least 1 AWS region."
#     }
# }

variable "azure" {
    type = object({
        subscription_id      = string
        client_id            = string
        client_secret        = string
        storage_account_name = string
        location             = string
    })
}

source "googlecompute" "gcp" {
    project_id              = "sourcegraph-ci"
    source_image_project_id = ["ubuntu-os-cloud"]
    source_image_family     = "ubuntu-2004-lts"
    disk_size               = 10
    ssh_username            = "packer"
    zone                    = "us-central1-c"
    disk_type               = "pd-ssd"
    image_name              = var.name
    image_description       = "Convenience image to run a docker registry pull-through cache. See github.com/sourcegraph/terraform-google-executors for how to use it."
    image_storage_locations = ["us"]
    tags                    = ["packer"]
    account_file            = "builder-sa-key.json"
}

source "amazon-ebs" "aws" {
    ami_name        = var.name
    ami_description = "Convenience image to run a docker registry pull-through cache. See github.com/sourcegraph/terraform-google-executors for how to use it."
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
    tags              = {
        Name          = var.name
        OS_Version    = "Ubuntu"
        Release       = "Latest"
        Base_AMI_Name = "{{ .SourceAMIName }}"
        Extra         = "{{ .SourceAMITags.TagName }}"
    }
}

source "azure-arm" "azure" {
    # sourcegraph/infrastructure/org/azure
    client_id       = var.azure.client_id
    client_secret   = var.azure.client_secret
    subscription_id = var.azure.subscription_id

    resource_group_name = "sourcegraph-ci"
    storage_account     = var.azure.storage_account_name # needs to be globally unique

    capture_container_name = "images"
    capture_name_prefix    = "packer"

    os_type         = "linux"
    image_publisher = "Canonical"
    image_offer     = "0001-com-ubuntu-server-focal"
    image_sku       = "20_04-lts"

    location = var.azure.location
    vm_size  = "Standard_A2_v2"
}

build {
    sources = [
        "source.googlecompute.gcp",
        "source.amazon-ebs.aws",
        "source.azure-arm.azure"
    ]

    provisioner "shell" {
        execute_command = "chmod +x {{ .Path }}; {{ .Vars }} sudo -E bash {{ .Path }}"
        script          = "install.sh"
        override        = {
            "gcp" = {
                environment_vars = ["PLATFORM_TYPE=gcp"]
            },
            "aws" = {
                environment_vars = ["PLATFORM_TYPE=aws"]
            },
            "azure" = {
                environment_vars = ["PLATFORM_TYPE=azure"]
            }
        }
    }
}
