#gem 'sufia-models'
# Returns an array containing the vhost 'CoSign service' value and URL
Sufia.config do |config|
  config.id_namespace = "und"
  config.fits_path = begin
    Rails.configuration.fits_path
  rescue NoMethodError
    "fits.sh"
  end
  config.fits_to_desc_mapping = begin
    Rails.configuration.fits_to_desc_mapping
  rescue NoMethodError => e
    { file_title: :title, file_author: :creator }
  end

  config.noid_template = '.reeddeeddedk'

  config.max_days_between_audits = 7

  config.cc_licenses = {
    'Attribution 3.0 United States' => 'http://creativecommons.org/licenses/by/3.0/us/',
    'Attribution-ShareAlike 3.0 United States' => 'http://creativecommons.org/licenses/by-sa/3.0/us/',
    'Attribution-NonCommercial 3.0 United States' => 'http://creativecommons.org/licenses/by-nc/3.0/us/',
    'Attribution-NoDerivs 3.0 United States' => 'http://creativecommons.org/licenses/by-nd/3.0/us/',
    'Attribution-NonCommercial-NoDerivs 3.0 United States' => 'http://creativecommons.org/licenses/by-nc-nd/3.0/us/',
    'Attribution-NonCommercial-ShareAlike 3.0 United States' => 'http://creativecommons.org/licenses/by-nc-sa/3.0/us/',
    'Public Domain Mark 1.0' => 'http://creativecommons.org/publicdomain/mark/1.0/',
    'CC0 1.0 Universal' => 'http://creativecommons.org/publicdomain/zero/1.0/',
    'All rights reserved' => 'All rights reserved'
  }

  config.cc_licenses_reverse = Hash[*config.cc_licenses.to_a.flatten.reverse]

  config.resource_types = {
    "Article" => "Article",
    "Audio" => "Audio",
    "Book" => "Book",
    "Capstone Project" => "Capstone Project",
    "Conference Proceeding" => "Conference Proceeding",
    "Dataset" => "Dataset",
    "Dissertation" => "Dissertation",
    "Image" => "Image",
    "Journal" => "Journal",
    "Map or Cartographic Material" => "Map or Cartographic Material",
    "Masters Thesis" => "Masters Thesis",
    "Part of Book" => "Part of Book",
    "Poster" => "Poster",
    "Presentation" => "Presentation",
    "Project" => "Project",
    "Report" => "Report",
    "Research Paper" => "Research Paper",
    "Software or Program Code" => "Software or Program Code",
    "Video" => "Video",
    "Other" => "Other",
  }

  config.permission_levels = {
    "Choose Access"=>"none",
    "View/Download" => "read",
    "Edit" => "edit"
  }

  config.owner_permission_levels = {
    "Edit" => "edit"
  }

  config.queue = Sufia::Resque::Queue

  # Map hostnames onto Google Analytics tracking IDs
  #config.google_analytics_id = 'UA-99999999-1'

end
