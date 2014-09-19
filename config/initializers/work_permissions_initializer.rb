WORK_PERMISSIONS = YAML.load( File.open( Rails.root.join( 'config/work_type_permissions.yml' ) ) ).freeze
