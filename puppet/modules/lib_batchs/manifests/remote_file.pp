define lib_batchs::remote_file($remote_location=undef, $mode='0644'){
  exec{"retrieve_${title}":
    command => "/usr/bin/wget -q ${remote_location} -O ${title}",
  }

  file{$title:
    mode    => $mode,
    replace => true,
    require => Exec["retrieve_${title}"],
  }
}
