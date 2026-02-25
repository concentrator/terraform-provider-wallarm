locals {
  hits = data.wallarm_hits.hits_test.hits
  hits_data_map = tomap({ for hit in local.hits : join(":", hit.id) => hit })
}

resource "local_file" "disable_stamp_rules" {
  for_each = toset(keys(local.hits_data_map))
  filename  = "${path.module}/disable_stamp_rules_${each.value}.tf"
  content  = templatefile("${path.module}/disable_stamp.tftpl", { hit = local.hits_data_map[each.value] })
  file_permission = "0644"
}

# resource "null_resource" "dependency_consumer" {
#   depends_on = [local_file.disable_stamp_rules]
  
#   provisioner "local-exec" {
#     command = "sleep 1 && terraform apply -lock=false -auto-approve"
#   }
# }

