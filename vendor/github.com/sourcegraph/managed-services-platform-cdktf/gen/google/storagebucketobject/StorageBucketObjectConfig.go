package storagebucketobject

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type StorageBucketObjectConfig struct {
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
	// The name of the containing bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#bucket StorageBucketObject#bucket}
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// The name of the object. If you're interpolating the name of this object, see output_name instead.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#name StorageBucketObject#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Cache-Control directive to specify caching behavior of object data.
	//
	// If omitted and object is accessible to all anonymous users, the default will be public, max-age=3600
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#cache_control StorageBucketObject#cache_control}
	CacheControl *string `field:"optional" json:"cacheControl" yaml:"cacheControl"`
	// Data as string to be uploaded.
	//
	// Must be defined if source is not. Note: The content field is marked as sensitive. To view the raw contents of the object, please define an output.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#content StorageBucketObject#content}
	Content *string `field:"optional" json:"content" yaml:"content"`
	// Content-Disposition of the object data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#content_disposition StorageBucketObject#content_disposition}
	ContentDisposition *string `field:"optional" json:"contentDisposition" yaml:"contentDisposition"`
	// Content-Encoding of the object data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#content_encoding StorageBucketObject#content_encoding}
	ContentEncoding *string `field:"optional" json:"contentEncoding" yaml:"contentEncoding"`
	// Content-Language of the object data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#content_language StorageBucketObject#content_language}
	ContentLanguage *string `field:"optional" json:"contentLanguage" yaml:"contentLanguage"`
	// Content-Type of the object data. Defaults to "application/octet-stream" or "text/plain; charset=utf-8".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#content_type StorageBucketObject#content_type}
	ContentType *string `field:"optional" json:"contentType" yaml:"contentType"`
	// customer_encryption block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#customer_encryption StorageBucketObject#customer_encryption}
	CustomerEncryption *StorageBucketObjectCustomerEncryption `field:"optional" json:"customerEncryption" yaml:"customerEncryption"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#detect_md5hash StorageBucketObject#detect_md5hash}.
	DetectMd5Hash *string `field:"optional" json:"detectMd5Hash" yaml:"detectMd5Hash"`
	// Whether an object is under event-based hold.
	//
	// Event-based hold is a way to retain objects until an event occurs, which is signified by the hold's release (i.e. this value is set to false). After being released (set to false), such objects will be subject to bucket-level retention (if any).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#event_based_hold StorageBucketObject#event_based_hold}
	EventBasedHold interface{} `field:"optional" json:"eventBasedHold" yaml:"eventBasedHold"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#id StorageBucketObject#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Resource name of the Cloud KMS key that will be used to encrypt the object.
	//
	// Overrides the object metadata's kmsKeyName value, if any.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#kms_key_name StorageBucketObject#kms_key_name}
	KmsKeyName *string `field:"optional" json:"kmsKeyName" yaml:"kmsKeyName"`
	// User-provided metadata, in key/value pairs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#metadata StorageBucketObject#metadata}
	Metadata *map[string]*string `field:"optional" json:"metadata" yaml:"metadata"`
	// retention block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#retention StorageBucketObject#retention}
	Retention *StorageBucketObjectRetention `field:"optional" json:"retention" yaml:"retention"`
	// A path to the data you want to upload. Must be defined if content is not.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#source StorageBucketObject#source}
	Source *string `field:"optional" json:"source" yaml:"source"`
	// The StorageClass of the new bucket object.
	//
	// Supported values include: MULTI_REGIONAL, REGIONAL, NEARLINE, COLDLINE, ARCHIVE. If not provided, this defaults to the bucket's default storage class or to a standard class.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#storage_class StorageBucketObject#storage_class}
	StorageClass *string `field:"optional" json:"storageClass" yaml:"storageClass"`
	// Whether an object is under temporary hold.
	//
	// While this flag is set to true, the object is protected against deletion and overwrites.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#temporary_hold StorageBucketObject#temporary_hold}
	TemporaryHold interface{} `field:"optional" json:"temporaryHold" yaml:"temporaryHold"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket_object#timeouts StorageBucketObject#timeouts}
	Timeouts *StorageBucketObjectTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

