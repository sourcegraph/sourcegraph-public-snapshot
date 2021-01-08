# Resources

#### YAML Standard Stuff

YAML does have anchors and references: https://blog.daemonl.com/2016/02/yaml.html
#### GitHub Actions

GitHub Actions allow setting the `output` of a step: https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-syntax-for-github-actions#jobsjob_idoutputs

And they do it by emitting a "magic line" in the steps:

```yaml
outputs:
  output1: ${{ steps.step1.outputs.test }}
  output2: ${{ steps.step2.outputs.test }}
steps:
- id: step1
  run: echo "::set-output name=test::hello"
- id: step2
  run: echo "::set-output name=test::world"
```

#### Azure DevOps Pipelines

Azure Pipelines support [templates](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/templates?view=azure-devops).

They also have [runtime parameters](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/runtime-parameters)

And steps can be grouped into [stages](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/stages?view=azure-devops&tabs=yaml)

#### HashiCorp Configuration Language (HCL)

Docs; [HashiCorp Configuration Language](https://www.terraform.io/docs/configuration/index.html)

HCL also has [output values](https://www.terraform.io/docs/configuration/outputs.html) and [input variables](https://www.terraform.io/docs/configuration/variables.html)

Interestingly enough, they do have "indexing" to specialize resources:

```hcl
resource "aws_subnet" "az" {
  # Create one subnet for each given availability zone.
  count = length(var.availability_zones)

  # For each subnet, use one of the specified availability zones.
  availability_zone = var.availability_zones[count.index]

  # By referencing the aws_vpc.main object, Terraform knows that the subnet
  # must be created only after the VPC is created.
  vpc_id = aws_vpc.main.id

  # Built-in functions and operators can be used for simple transformations of
  # values, such as computing a subnet address. Here we create a /20 prefix for
  # each subnet, using consecutive addresses for each availability zone,
  # such as 10.1.16.0/20 .
  cidr_block = cidrsubnet(aws_vpc.main.cidr_block, 4, count.index+1)
}
```

How is this even evaluated, though? They _set_ `count` to the length of another var, but also use `count.index`? Does writing to `count` overwrite the count for the resource and then `count.index` is a "getter" that accesses the current iteration?
