/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import React from 'react';
import { Box, Flex, Heading, Text } from 'pouncejs';
import EmptyNotepadImg from 'Assets/illustrations/empty-notepad.svg';
import RuleCreateButton from '../CreateButton';

const ListRulesPageEmptyDataFallback: React.FC = () => {
  return (
    <Flex justify="center" align="center" direction="column">
      <Box my={10}>
        <img alt="Empty Notepad illustration" src={EmptyNotepadImg} width="auto" height={300} />
      </Box>
      <Heading mb={6}>No rules found</Heading>
      <Text color="gray-300" textAlign="center" mb={8}>
        Writing rules will allow you to get alerts about suspicious activity in your system
      </Text>
      <RuleCreateButton />
    </Flex>
  );
};

export default ListRulesPageEmptyDataFallback;
