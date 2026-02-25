# resource "wallarm_global_mode" "global_block" {
#   filtration_mode = "block"
#   rechecker_mode = "on"
# }


module "fp" {
  source = "./fp"
}

