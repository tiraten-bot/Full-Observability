variable "cluster_name" {
  type = string
}

variable "cluster_version" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "node_groups" {
  type = map(object({
    desired_capacity = number
    max_capacity     = number
    min_capacity     = number
    instance_types   = list(string)
    capacity_type    = string
    labels           = optional(map(string), {})
    taints           = optional(list(object({
      key    = string
      value  = string
      effect = string
    })), [])
    tags = optional(map(string), {})
  }))
}

variable "cluster_endpoint_public_access_cidrs" {
  type    = list(string)
  default = ["0.0.0.0/0"]
}

variable "kms_key_arn" {
  type    = string
  default = ""
}

variable "vpc_cni_version" {
  type    = string
  default = null
}

variable "coredns_version" {
  type    = string
  default = null
}

variable "kube_proxy_version" {
  type    = string
  default = null
}

variable "tags" {
  type = map(string)
}

