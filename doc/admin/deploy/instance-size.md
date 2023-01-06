---
title: Instance Size
---

# Instance Size

You should configure your Sourcegraph instance based on the number of users and repositories you have using our size chart shown below.

## Size chart

If you fall between two sizes, choose the larger of the two. For examples:

1. If you have 8,000 users with 80,000 repositories, your instance size would be **L**. 
2. If you have 1,000 users with 80,000 repositories, your instance size would still be **L**. 

|                  | **XS**        | **S**          | **M**          | **L**          | **XL**         |
|:-----------------|:-------------:|:--------------:|:--------------:|:--------------:|:--------------:|
| **Users**        | Up to 500     | Up to 1,000    | Up to 5,000    | Up to 10,000   | Up to 20,000   |
| **Repositories** | Up to 5,000   | Up to 10,000   | Up to 50,000   | Up to 100,000  | Up to 250,000  |
| **vCPU**         | 8             | 16             | 32             | 48             | 96             |
| **Memory (GB)**  | 32            | 64             | 128            | 192            | 384            |
| **SSD Required** | Yes           | Yes            | Yes            | Yes            | Yes            |

## Instance type

We recommend the following instance type for the cloud providera listed below.

|                  | **XS**        | **S**          | **M**          | **L**          | **XL**         |
|:-----------------|:-------------:|:--------------:|:--------------:|:--------------:|:--------------:|
| **AWS**          | m6a.2xlarge   | m6a.4xlarge    | m6a.8xlarge    | m6a.12xlarge   | m6a.24xlarge   |
| **Azure**        | D8_v3         | D16_v3         | D32_v3         | D48_v3         | D64_v3         |
| **GCP**          | n2-standard-8 | n2-standard-16 | n2-standard-32 | n2-standard-48 | n2-standard-96 |

