full_name = input_data.get('full_name', '')
email = input_data.get('email', '')
site_id = input_data.get('site_id', '')
rating = input_data.get('rating', '')
use_cases = input_data.get('use_cases', [])
other_use_case = input_data.get('other_use_case', '')
additional_information = input_data.get('additional_information', '')

#why_rating = input_data.get('why_rating', '')
#better_product = input_data.get('better_product', '')

description = ""
readable_use_cases = ""
notify = None
create_productboard_note = False


# TODO: Notify from use cases?
# def notify(why, better):
#    """ Lazily match one PM if feedback matches a product area"""
#    if one_match(['campaign','batch'],why,better):
#        return "@malo"
#    elif one_match(['hover','intel'],why,better):
#        return "@maria"
#    elif one_match(['insight'],why,better):
#        return "@Joel Kwartler"
#    else:
#        return None
#
#
# def one_match(words,why,better):
#     for word in words:
#         if why and word in why:
#             return True
#         elif better and word in better:
#             return True
#     return False

def get_readable_use_cases(use_cases):
    use_case_dict = {
        "UNDERSTAND_NEW_CODE": "Understand a new codebase",
        "FIX_SECURITY_VULNERABILITIES": "Fix security vulnerabilities",
        "REUSE_CODE": "Reuse code",
        "RESPOND_TO_INCIDENTS": "Respond to incidents",
        "IMPROVE_CODE_QUALITY": "Improve code quality"
    }
    return ", ".join([use_case_dict[use_case] for use_case in use_cases])

def is_feedback(s):
    """Filter meaningful feedback, that contains at least three words"""
    return len([x for x in s.split(" ") if x != ''])>3

if full_name:
    description += "*Name:* " + full_name + "\n"
if email:
    description += "*Email:* " + email + "\n"
description += "*SiteId:* " + site_id + "\n"
description += "*Score:* " + rating + "\n"
if use_cases:
    description += "*Use cases:*\n>" + use_cases + "\n"
    readable_use_cases = get_readable_use_cases(use_cases)
if other_use_case:
    description += "*What else are you using Sourcegraph to do?*\n>" + other_use_case + "\n"
if additional_information:
    description += "*Anything else you'd like to share with us?:*\n>" + additional_information + "\n"
if additional_information and is_feedback(additional_information):
    create_productboard_note = False # TODO Set to True
# notify = notify(why_rating, better_product)

return {"description": description, "create_productboard_note": create_productboard_note, "readable_use_cases": readable_use_cases}
