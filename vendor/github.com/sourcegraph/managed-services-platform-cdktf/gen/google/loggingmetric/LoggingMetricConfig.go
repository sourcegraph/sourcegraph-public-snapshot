package loggingmetric

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type LoggingMetricConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// An advanced logs filter (https://cloud.google.com/logging/docs/view/advanced-filters) which is used to match log entries.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#filter LoggingMetric#filter}
	Filter *string `field:"required" json:"filter" yaml:"filter"`
	// The client-assigned metric identifier.
	//
	// Examples - "error_count", "nginx/requests".
	// Metric identifiers are limited to 100 characters and can include only the following
	// characters A-Z, a-z, 0-9, and the special characters _-.,+!*',()%/. The forward-slash
	// character (/) denotes a hierarchy of name pieces, and it cannot be the first character
	// of the name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#name LoggingMetric#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The resource name of the Log Bucket that owns the Log Metric.
	//
	// Only Log Buckets in projects
	// are supported. The bucket has to be in the same project as the metric.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#bucket_name LoggingMetric#bucket_name}
	BucketName *string `field:"optional" json:"bucketName" yaml:"bucketName"`
	// bucket_options block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#bucket_options LoggingMetric#bucket_options}
	BucketOptions *LoggingMetricBucketOptions `field:"optional" json:"bucketOptions" yaml:"bucketOptions"`
	// A description of this metric, which is used in documentation. The maximum length of the description is 8000 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#description LoggingMetric#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// If set to True, then this metric is disabled and it does not generate any points.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#disabled LoggingMetric#disabled}
	Disabled interface{} `field:"optional" json:"disabled" yaml:"disabled"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#id LoggingMetric#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// A map from a label key string to an extractor expression which is used to extract data from a log entry field and assign as the label value.
	//
	// Each label key specified in the LabelDescriptor must
	// have an associated extractor expression in this map. The syntax of the extractor expression is
	// the same as for the valueExtractor field.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#label_extractors LoggingMetric#label_extractors}
	LabelExtractors *map[string]*string `field:"optional" json:"labelExtractors" yaml:"labelExtractors"`
	// metric_descriptor block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#metric_descriptor LoggingMetric#metric_descriptor}
	MetricDescriptor *LoggingMetricMetricDescriptor `field:"optional" json:"metricDescriptor" yaml:"metricDescriptor"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#project LoggingMetric#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#timeouts LoggingMetric#timeouts}
	Timeouts *LoggingMetricTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// A valueExtractor is required when using a distribution logs-based metric to extract the values to record from a log entry.
	//
	// Two functions are supported for value extraction - EXTRACT(field) or
	// REGEXP_EXTRACT(field, regex). The argument are 1. field - The name of the log entry field from which
	// the value is to be extracted. 2. regex - A regular expression using the Google RE2 syntax
	// (https://github.com/google/re2/wiki/Syntax) with a single capture group to extract data from the specified
	// log entry field. The value of the field is converted to a string before applying the regex. It is an
	// error to specify a regex that does not include exactly one capture group.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#value_extractor LoggingMetric#value_extractor}
	ValueExtractor *string `field:"optional" json:"valueExtractor" yaml:"valueExtractor"`
}

