from invoke import run, Responder
import os
import json
import pytest

this_test_path = os.path.dirname(os.path.realpath(__file__))
# Calculate absolute path of the CLI
cli_path = os.path.join(this_test_path, '..', 'arduino-cli')

# Useful reference: 
# http://docs.pyinvoke.org/en/1.2/api/runners.html#invoke.runners.Result


def cli_line(*args):
    # Accept a list of arguments cli_line('lib list --format json')
    # Return a full command line string e.g. 'arduino-cli help --format json'
    cli_full_line = ' '.join([cli_path, ' '.join(str(arg) for arg in args)])
    return cli_full_line


def run_command(*args):
    result = run(cli_line(*args), echo=False, hide=True)
    return result


def test_command_help():
    result = run_command('help')
    assert result.ok
    assert result.stderr == ''
    assert 'Usage' in result.stdout


def test_command_lib_list():
    result = run_command('lib list')
    assert result.stderr == ''
    result = run_command('lib list', '--format json')
    assert '{}' == result.stdout


def test_command_lib_install():
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    # Should be safe to run install multiple times
    result_1 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_1.ok
    result_2 = run_command('lib install {}'.format(' '.join(libs)))
    assert result_2.ok


def test_command_lib_remove():
    libs = ['\"AzureIoTProtocol_MQTT\"', '\"CMMC MQTT Connector\"', '\"WiFiNINA\"']
    result = run_command('lib uninstall {}'.format(' '.join(libs)))
    assert result.ok

@pytest.mark.slow
def test_command_lib_search():
    result = run_command('lib search')
    out_lines = result.stdout.splitlines()
    libs = []
    # Create an array with just the name of the vars
    for line in out_lines:
        if 'Name: ' in line:
            libs.append(line.split()[1].strip('\"'))
    number_of_libs = len(libs)
    # It would be strange to have less than 2000 Arduino Libs published
    assert number_of_libs > 2000
    result = run_command('lib search --format json')
    libs_found_from_json = json.loads(result.stdout)
    number_of_libs_from_json = len(libs_found_from_json.get('libraries'))
    assert number_of_libs == number_of_libs_from_json


def test_command_board_list():
    result = run_command('board list --format json')
    # check is a valid json and contains a list of ports
    ports = json.loads(result.stdout).get('ports')
    assert isinstance(ports, list)
    for port in ports:
        assert 'protocol' in port
        assert 'protocol_label' in port


def test_command_board_listall():
    result = run_command('board listall')
    assert ['Board', 'Name', 'FQBN'] == result.stdout.splitlines()[0].strip().split()

def test_command_version():
    result = run_command('version --format json')
    parsed_out = json.loads(result.stdout)

    assert parsed_out.get('command', False) == 'arduino-cli'
    assert parsed_out.get('version', False)
    assert parsed_out.get('commit', False)
    assert parsed_out.get('build_date', False)
